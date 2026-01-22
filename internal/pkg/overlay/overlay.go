package overlay

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"text/template/parse"

	"github.com/Masterminds/sprig/v3"
	"github.com/coreos/go-systemd/v22/unit"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var ErrDoesNotExist = fmt.Errorf("overlay does not exist")

// Overlay represents an overlay directory path.
type Overlay string

// Name returns the base name of the overlay directory.
//
// This is derived from the full path of the overlay.
func (overlay Overlay) Name() string {
	return path.Base(overlay.Path())
}

// Path returns the string representation of the overlay path.
//
// This method allows the Overlay type to be easily converted back to its
// underlying string representation.
func (overlay Overlay) Path() string {
	return string(overlay)
}

// Rootfs returns the path to the root filesystem (rootfs) within the overlay.
//
// If the "rootfs" directory exists inside the overlay path, it returns the
// path to the "rootfs" directory. Otherwise, it checks if the overlay path
// itself is a directory and returns that. If neither exists, it defaults to
// returning the "rootfs" path.
func (overlay Overlay) Rootfs() string {
	rootfs := path.Join(overlay.Path(), "rootfs")
	if util.IsDir(rootfs) {
		return rootfs
	} else if util.IsDir(overlay.Path()) {
		return overlay.Path()
	} else {
		return rootfs
	}
}

// File constructs a full path to a file within the overlay's root filesystem.
//
// Parameters:
//   - filePath: The relative path of the file within the overlay.
//
// Returns:
//   - The full path to the specified file in the overlay's rootfs.
//     If the specified path is not contained within the overlay, the empty string is returned.
func (overlay Overlay) File(filePath string) string {
	rootfs := overlay.Rootfs()
	fullPath := path.Join(rootfs, filePath)
	cleanPath := filepath.Clean(fullPath)
	cleanRootfs := filepath.Clean(rootfs)
	rel, err := filepath.Rel(cleanRootfs, cleanPath)
	if err != nil {
		return ""
	}

	if strings.HasPrefix(rel, "..") {
		return ""
	}

	return cleanPath
}

// Exists checks whether the overlay path exists and is a directory.
//
// Returns:
//   - true if the overlay path exists and is a directory; false otherwise.
func (overlay Overlay) Exists() bool {
	return util.IsDir(overlay.Path())
}

// IsSiteOverlay determines whether the overlay is a site overlay.
//
// A site overlay is identified by its parent directory matching the configured
// site overlay directory path.
//
// Returns:
//   - true if the overlay is a site overlay; false otherwise.
func (overlay Overlay) IsSiteOverlay() bool {
	siteDir := filepath.Clean(config.Get().Paths.SiteOverlaydir())
	overlayPath := filepath.Clean(overlay.Path())
	if rel, err := filepath.Rel(siteDir, overlayPath); err != nil {
		return false
	} else {
		return !strings.HasPrefix(rel, "..")
	}
}

// IsDistributionOverlay determines whether the overlay is a distribution overlay.
//
// A distribution overlay is identified by its parent directory matching the configured
// distribution overlay directory path.
//
// Returns:
//   - true if the overlay is a distribution overlay; false otherwise.
func (overlay Overlay) IsDistributionOverlay() bool {
	siteDir := filepath.Clean(config.Get().Paths.DistributionOverlaydir())
	overlayPath := filepath.Clean(overlay.Path())
	if rel, err := filepath.Rel(siteDir, overlayPath); err != nil {
		return false
	} else {
		return !strings.HasPrefix(rel, "..")
	}
}

func (overlay Overlay) AddFile(filePath string, content []byte, parents bool, force bool) error {
	wwlog.Info("Creating file %s in overlay %s, force: %v", filePath, overlay.Name(), force)

	if !overlay.IsSiteOverlay() {
		siteOverlay, err := overlay.CloneToSite()
		if err != nil {
			return fmt.Errorf("failed to clone distribution overlay '%s' to site overlay: %w", overlay.Name(), err)
		}
		// replace the overlay with newly created siteOverlay
		overlay = siteOverlay
	}
	fullPath := overlay.File(filePath)
	// create necessary parent directories
	if parents {
		if err := os.MkdirAll(path.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("failed to create parent directories for %s: %w", fullPath, err)
		}
	}

	// if the file already exists and force is false, return an error
	if util.IsFile(fullPath) {
		if force {
			return os.WriteFile(fullPath, content, 0o644)
		}
		return fmt.Errorf("file %s already exists in overlay %s", filePath, overlay.Name())
	}

	return os.WriteFile(fullPath, content, 0o644)
}

func (overlay Overlay) Delete(force bool) (err error) {
	wwlog.Info("Deleting overlay %s, force: %v", overlay.Name(), force)
	if overlay.IsDistributionOverlay() {
		return fmt.Errorf("cannot delete a distribution overlay: %s", overlay.Name())
	}
	if force {
		err := os.RemoveAll(overlay.Path())
		if err != nil {
			return fmt.Errorf("failed to delete overlay forcely: %w", err)
		}
	} else {
		// remove rootfs at first
		if err = os.Remove(overlay.Rootfs()); err != nil {
			return fmt.Errorf("failed to delete overlay: %w", err)
		}
		if overlay.Exists() {
			if err = os.Remove(overlay.Path()); err != nil {
				return fmt.Errorf("failed to delete overlay: %w", err)
			}
		}
	}
	return nil
}

// DeleteFile deletes a file or the entire overlay directory.
// before deletion.
func (overlay Overlay) DeleteFile(filePath string, force, cleanup bool) (err error) {
	wwlog.Info("Deleting file %s from overlay %s, force: %v, cleanup: %v", filePath, overlay.Name(), force, cleanup)
	// first check if file exists
	if !util.IsFile(overlay.File(filePath)) {
		return fmt.Errorf("file %s does not exist in overlay %s", filePath, overlay.Name())
	}
	if overlay.IsDistributionOverlay() {
		siteOverlay, err := overlay.CloneToSite()
		if err != nil {
			return fmt.Errorf("failed to clone distribution overlay '%s' to site overlay: %w", overlay.Name(), err)
		}
		// replace the overlay with newly created siteOverlay
		overlay = siteOverlay
	}
	fullPath := overlay.File(filePath)
	if force {
		if err := os.RemoveAll(fullPath); err != nil {
			return fmt.Errorf("failed to delete file %s from overlay %s: %w", filePath, overlay.Name(), err)
		}
	} else {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to delete file %s from overlay %s: %w", filePath, overlay.Name(), err)
		}
	}

	if cleanup {
		// cleanup the empty parents
		i := path.Dir(fullPath)
		for i != overlay.Rootfs() {
			wwlog.Debug("Evaluating directory to remove: %s", i)
			err := os.Remove(i)
			if err != nil {
				// if the directory is not empty, we stop here
				if !os.IsNotExist(err) {
					wwlog.Debug("Could not remove directory %s: %v", i, err)
				}
				break
			}
			wwlog.Debug("Removed empty directory: %s", i)
			i = path.Dir(i)
		}
	}
	return nil
}

// chmod for the given path in the overlay
func (overlay Overlay) Chmod(path string, mode uint64) (err error) {
	if !overlay.IsSiteOverlay() {
		overlay, err = overlay.CloneToSite()
		if err != nil {
			return err
		}
	}
	fullPath := overlay.File(path)
	if !(util.IsFile(fullPath) || util.IsDir(fullPath)) {
		return fmt.Errorf("file does not exist within overlay: %s:%s", overlay.Name(), fullPath)
	}

	return os.Chmod(fullPath, os.FileMode(mode))
}

// chown file or dir in overlay
func (overlay Overlay) Chown(path string, uid, gid int) (err error) {
	if !overlay.IsSiteOverlay() {
		overlay, err = overlay.CloneToSite()
		if err != nil {
			return err
		}
	}
	fullPath := overlay.File(path)
	if !(util.IsFile(fullPath) || util.IsDir(fullPath)) {
		return fmt.Errorf("file does not exist within overlay: %s:%s", overlay.Name(), fullPath)
	}
	return os.Chown(fullPath, uid, gid)
}

func (overlay Overlay) Mkdir(path string, mode int32) (err error) {
	if !overlay.IsSiteOverlay() {
		overlay, err = overlay.CloneToSite()
		if err != nil {
			return err
		}
	}
	fullPath := overlay.File(path)
	if util.IsFile(fullPath) || util.IsDir(fullPath) {
		wwlog.Warn("path already exists, overwriting permissions: %s:%s", overlay.Name(), fullPath)
	}
	return os.MkdirAll(fullPath, os.FileMode(mode))
}

// FieldInfo contains detailed type information about a template variable
type FieldInfo struct {
	Field      reflect.StructField // Complete field metadata including tags
	ParentType reflect.Type        // The containing struct type
	FullPath   string              // Full path like ".NetDevs.Ipaddr6"
	VarName    string              // Original variable name from template
}

// ParseVarFields returns detailed type information for each variable in the template
// by using reflection to resolve the actual struct fields being referenced.
func (overlay Overlay) ParseVarFields(file string) map[string]FieldInfo {
	if !strings.HasSuffix(file, ".ww") {
		return nil
	}
	fullPath := overlay.File(file)
	if !util.IsFile(fullPath) {
		wwlog.Error("Template file does not exist in overlay %s: %s", overlay.Name(), file)
		return nil
	}

	funcMap, _, _ := getTemplateFuncMap(fullPath, TemplateStruct{})
	tmpl, err := template.New(path.Base(fullPath)).Option("missingkey=default").Funcs(funcMap).ParseFiles(fullPath)
	if err != nil {
		wwlog.Error("Could not parse template file %s: %s", fullPath, err)
		return nil
	}

	result := make(map[string]FieldInfo)
	rootType := reflect.TypeOf(TemplateStruct{})

	// Track range variables and their types
	rangeVars := make(map[string]reflect.Type)
	// Initialize $ to refer to the root template context
	rangeVars["$"] = rootType

	if tmpl.Tree != nil && tmpl.Tree.Root != nil {
		walkParseTree(tmpl.Tree.Root, rootType, "", rangeVars, result)
	}

	return result
}

// ParseCommentVars parses a template file for comments that contain variable documentations.
// The comments must be in the format `{{/* key: value */}}`. The content is parsed as YAML.
func (overlay Overlay) ParseCommentVars(file string) (retMap map[string]string) {
	retMap = make(map[string]string)
	if !strings.HasSuffix(file, ".ww") {
		return nil
	}
	fullPath := overlay.File(file)
	if !util.IsFile(fullPath) {
		wwlog.Error("Template file does not exist in overlay %s: %s", overlay.Name(), file)
		return nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		wwlog.Error("Could not read template file %s: %s", fullPath, err)
		return nil
	}

	re := regexp.MustCompile(`{{\s*/\*\s*(.*?):\s*(.*?)\s*\*/\s*}}`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	if len(matches) > 0 {
		wwlog.Debug("matches: %v len(%d:%d)", matches, len(matches), len(matches[0]))
	} else {
		wwlog.Debug("matches: [] len(0)")
	}
	for i := range matches {
		if len(matches[i]) > 2 {
			retMap[matches[i][1]] = matches[i][2]
		}
	}
	return
}

// walkParseTree recursively traverses the template's parse tree and resolves
// variable references to actual struct fields using reflection.
func walkParseTree(node parse.Node, currentType reflect.Type, currentPath string, rangeVars map[string]reflect.Type, result map[string]FieldInfo) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parse.ActionNode:
		// Handle variable assignments like $var := expr
		if n.Pipe != nil && len(n.Pipe.Decl) > 0 && len(n.Pipe.Cmds) > 0 {
			// Try to resolve the type of the expression being assigned
			cmd := n.Pipe.Cmds[0]
			if len(cmd.Args) > 0 {
				// Check for field access (e.g., $NetDevs := .NetDevs)
				if field, ok := cmd.Args[0].(*parse.FieldNode); ok {
					fieldInfo := resolveFieldChain(currentType, field.Ident, currentPath)
					if fieldInfo != nil {
						// Record all declared variables with this type
						for _, decl := range n.Pipe.Decl {
							rangeVars[decl.Ident[0]] = fieldInfo.Field.Type
						}
					}
				} else if varNode, ok := cmd.Args[0].(*parse.VariableNode); ok {
					// Check for variable-to-variable assignment (e.g., $b := $a)
					if len(varNode.Ident) == 1 {
						// Simple variable reference
						varBaseName := varNode.Ident[0]
						if varType, exists := rangeVars[varBaseName]; exists {
							for _, decl := range n.Pipe.Decl {
								rangeVars[decl.Ident[0]] = varType
							}
						}
					}
				}
			}
		}
		walkParseTree(n.Pipe, currentType, currentPath, rangeVars, result)

	case *parse.IfNode:
		walkParseTree(n.Pipe, currentType, currentPath, rangeVars, result)
		walkParseTree(n.List, currentType, currentPath, rangeVars, result)
		walkParseTree(n.ElseList, currentType, currentPath, rangeVars, result)

	case *parse.ListNode:
		if n != nil {
			for _, child := range n.Nodes {
				walkParseTree(child, currentType, currentPath, rangeVars, result)
			}
		}

	case *parse.RangeNode:
		// Handle range statements like {{range $key, $val := .NetDevs}}
		// First, walk the pipe to record the variable being ranged over
		walkParseTree(n.Pipe, currentType, currentPath, rangeVars, result)

		if n.Pipe != nil && len(n.Pipe.Cmds) > 0 {
			// Get what's being ranged over to determine element type
			var fieldInfo *FieldInfo

			// Check for both FieldNode (e.g., .NetDevs) and VariableNode (e.g., $disk.PartitionList)
			rangeField := extractFieldFromPipe(n.Pipe)
			if rangeField != nil {
				// Resolve the type of the field being ranged over
				fieldInfo = resolveFieldChain(currentType, rangeField.Ident, currentPath)
			} else {
				// Check if it's a variable access
				for _, cmd := range n.Pipe.Cmds {
					for _, arg := range cmd.Args {
						if varNode, ok := arg.(*parse.VariableNode); ok {
							varBaseName := varNode.Ident[0]
							if varType, exists := rangeVars[varBaseName]; exists {
								if len(varNode.Ident) > 1 {
									// Multi-part variable like $disk.PartitionList
									fieldIdents := varNode.Ident[1:] // Skip the variable name
									fieldInfo = resolveFieldChain(varType, fieldIdents, currentPath)
								} else {
									// Simple variable like $NetDevs
									fieldInfo = &FieldInfo{
										Field: reflect.StructField{
											Name: varBaseName,
											Type: varType,
										},
										ParentType: currentType,
										FullPath:   currentPath,
									}
								}
								break
							}
						}
					}
					if fieldInfo != nil {
						break
					}
				}
			}

			if fieldInfo != nil {
				rangeType := fieldInfo.Field.Type

				// For slices and arrays, get the element type
				if rangeType.Kind() == reflect.Slice || rangeType.Kind() == reflect.Array {
					rangeType = rangeType.Elem()
				}

				// For maps, get the value type
				if rangeType.Kind() == reflect.Map {
					rangeType = rangeType.Elem()
				}

				// Dereference pointers
				if rangeType.Kind() == reflect.Ptr {
					rangeType = rangeType.Elem()
				}

				// Track the range variable assignments
				if len(n.Pipe.Decl) > 0 {
					if len(n.Pipe.Decl) == 2 {
						// Two-variable range: $key, $val := .Map or $index, $val := .Slice
						keyType := reflect.TypeOf(0) // Default to int for slice indices
						if fieldInfo.Field.Type.Kind() == reflect.Map {
							keyType = fieldInfo.Field.Type.Key()
						}
						rangeVars[n.Pipe.Decl[0].Ident[0]] = keyType   // First variable is key/index
						rangeVars[n.Pipe.Decl[1].Ident[0]] = rangeType // Second variable is value
					} else {
						// Single-variable range: $val := .Slice
						rangeVars[n.Pipe.Decl[0].Ident[0]] = rangeType
					}
				}

				// Walk the range body with the element type
				walkParseTree(n.List, rangeType, fieldInfo.FullPath, rangeVars, result)
			}
		}
		walkParseTree(n.ElseList, currentType, currentPath, rangeVars, result)

	case *parse.WithNode:
		walkParseTree(n.Pipe, currentType, currentPath, rangeVars, result)
		walkParseTree(n.List, currentType, currentPath, rangeVars, result)
		walkParseTree(n.ElseList, currentType, currentPath, rangeVars, result)

	case *parse.TemplateNode:
		walkParseTree(n.Pipe, currentType, currentPath, rangeVars, result)

	case *parse.PipeNode:
		if n != nil {
			for _, cmd := range n.Cmds {
				for _, arg := range cmd.Args {
					walkParseTree(arg, currentType, currentPath, rangeVars, result)
				}
			}
		}

	case *parse.FieldNode:
		// Field access like .Ipmi.Ipaddr or $netdev.Ipaddr
		varName := n.String()

		// Determine the base type for field resolution
		baseType := currentType
		basePath := currentPath

		// Check if this is a variable access (starts with $)
		if strings.HasPrefix(varName, "$") {
			// Extract variable name (e.g., "$netdev.Device" -> "netdev")
			parts := strings.SplitN(varName[1:], ".", 2)
			if len(parts) > 0 {
				varBaseName := parts[0]
				if varType, exists := rangeVars[varBaseName]; exists {
					baseType = varType
					basePath = currentPath
				}
			}
		}

		fieldInfo := resolveFieldChain(baseType, n.Ident, basePath)

		// If resolution failed, try adding "P" suffix to last identifier
		// This handles methods like Primary() backed by PrimaryP field
		if fieldInfo == nil && len(n.Ident) >= 1 {
			identWithP := make([]string, len(n.Ident))
			copy(identWithP, n.Ident)
			identWithP[len(identWithP)-1] += "P"
			fieldInfo = resolveFieldChain(baseType, identWithP, basePath)
			// Use the original variable name (without P)
			if fieldInfo != nil {
				fieldInfo.VarName = varName
			}
		}

		if fieldInfo != nil {
			if fieldInfo.VarName == "" {
				fieldInfo.VarName = varName
			}
			result[varName] = *fieldInfo
		}

	case *parse.VariableNode:
		// Variable reference like $netdev or possibly $netdev.Field
		varName := n.String()

		// Check if this is a simple variable or a field access on a variable
		if len(n.Ident) > 1 {
			// This is $var.Field or $var.Field.Method - handle like a FieldNode
			varBaseName := n.Ident[0]
			if varType, exists := rangeVars[varBaseName]; exists {
				// Resolve the field chain starting from the variable's type
				fieldIdents := n.Ident[1:] // Skip the variable name, keep the fields
				fieldInfo := resolveFieldChain(varType, fieldIdents, currentPath)

				// If resolution failed and we have multiple parts, the last part might be a method
				// Try resolving without the last identifier (e.g., OnBoot.BoolDefaultTrue -> OnBoot)
				if fieldInfo == nil && len(fieldIdents) > 1 {
					fieldIdents = fieldIdents[:len(fieldIdents)-1]
					fieldInfo = resolveFieldChain(varType, fieldIdents, currentPath)
					// Use a simplified variable name without the method
					if fieldInfo != nil {
						// Build simplified name: $varBaseName.field1.field2 (without method)
						simplifiedName := varBaseName
						for _, ident := range fieldIdents {
							simplifiedName += "." + ident
						}
						fieldInfo.VarName = simplifiedName
					}
				}

				// If resolution still failed, try adding "P" suffix to last identifier
				// This handles methods like Primary() backed by PrimaryP field
				if fieldInfo == nil && len(fieldIdents) >= 1 {
					// Try with "P" suffix on the last identifier
					fieldIdentsWithP := make([]string, len(fieldIdents))
					copy(fieldIdentsWithP, fieldIdents)
					fieldIdentsWithP[len(fieldIdentsWithP)-1] += "P"
					fieldInfo = resolveFieldChain(varType, fieldIdentsWithP, currentPath)
					// Use the original variable name (without P)
					if fieldInfo != nil {
						fieldInfo.VarName = varName
					}
				}

				if fieldInfo != nil {
					if fieldInfo.VarName == "" {
						fieldInfo.VarName = varName
					}
					result[fieldInfo.VarName] = *fieldInfo
				}
			}
		} else if len(n.Ident) == 1 {
			// Simple variable reference
			varBaseName := n.Ident[0]
			if varType, exists := rangeVars[varBaseName]; exists {
				result[varName] = FieldInfo{
					VarName:    varName,
					ParentType: varType,
					FullPath:   currentPath,
				}
			}
		}
	}
}

// extractFieldFromPipe extracts the FieldNode from a pipe (used in range statements)
func extractFieldFromPipe(pipe *parse.PipeNode) *parse.FieldNode {
	if pipe == nil || len(pipe.Cmds) == 0 {
		return nil
	}
	for _, cmd := range pipe.Cmds {
		for _, arg := range cmd.Args {
			if field, ok := arg.(*parse.FieldNode); ok {
				return field
			}
		}
	}
	return nil
}

// resolveFieldChain walks a chain of field identifiers (like ["Ipmi", "Ipaddr"])
// and returns the final field's information using reflection.
// Methods are resolved and reported. If a method has a backing field with "P" suffix,
// the backing field's metadata is used; otherwise, the method's return type is used.
func resolveFieldChain(rootType reflect.Type, idents []string, basePath string) *FieldInfo {
	if rootType.Kind() == reflect.Ptr {
		rootType = rootType.Elem()
	}

	if len(idents) == 0 {
		return nil
	}

	currentType := rootType
	fullPath := basePath
	var finalField reflect.StructField
	var parentType reflect.Type

	for i, fieldName := range idents {
		if currentType.Kind() != reflect.Struct {
			return nil
		}

		field, found := currentType.FieldByName(fieldName)
		if !found {
			// Field not found - try to find a method with this name
			// This handles methods like DiskList(), PartitionList(), Id(), ShouldExist()
			ptrType := reflect.PointerTo(currentType)
			method, methodFound := ptrType.MethodByName(fieldName)
			if !methodFound {
				return nil
			}

			// Get the method's return type (first return value)
			methodType := method.Type
			if methodType.NumOut() == 0 {
				return nil
			}
			returnType := methodType.Out(0)

			// Check if there's a backing field with "P" suffix
			// Only methods with backing fields should be reported as user-facing variables
			backingFieldName := fieldName + "P"
			backingField, backingFound := currentType.FieldByName(backingFieldName)

			// For the last identifier in the chain
			if i == len(idents)-1 {
				if backingFound {
					// Use the backing field's metadata (tags) but the method's return type
					// This gives us the documentation from ShouldExistP but the type from ShouldExist()
					finalField = reflect.StructField{
						Name: fieldName,
						Type: returnType, // Use method's return type, not backing field's type
						Tag:  backingField.Tag,
					}
					parentType = currentType
					fullPath += "." + fieldName // Use method name in path, not field name
					currentType = returnType
				} else {
					// No backing field - create a field with the method's return type
					// This allows documenting methods via comments
					finalField = reflect.StructField{
						Name: fieldName,
						Type: returnType,
					}
					parentType = currentType
					fullPath += "." + fieldName
					currentType = returnType
				}
			} else {
				// Not the last identifier - we're in the middle of a chain (e.g., DiskList in node.DiskList.PartitionList)
				// Continue walking with the method's return type for type resolution
				parentType = currentType
				fullPath += "." + fieldName
				currentType = returnType

				// We don't have a real field yet, just continue to next identifier
				// Don't set finalField here as we need to keep walking
				continue
			}
		} else {
			// Field found
			finalField = field
			parentType = currentType
			fullPath += "." + fieldName
			currentType = field.Type
		}

		// Dereference pointer types for next iteration
		if currentType.Kind() == reflect.Ptr {
			currentType = currentType.Elem()
		}

		// For map types, remaining identifiers are map keys, not fields
		if currentType.Kind() == reflect.Map && i < len(idents)-1 {
			// Append remaining path as map keys
			mapKeyPath := ""
			for j := i + 1; j < len(idents); j++ {
				mapKeyPath += "." + idents[j]
			}
			fullPath += mapKeyPath

			// Create a synthetic field with the map's value type
			// For Tags (map[string]string), accessing Tags.key should return string, not map[string]string
			valueType := currentType.Elem()
			return &FieldInfo{
				Field: reflect.StructField{
					Name: idents[len(idents)-1], // Use the last key as the field name
					Type: valueType,             // Use the map's value type
				},
				ParentType: parentType,
				FullPath:   fullPath,
			}
		}
	}

	return &FieldInfo{
		Field:      finalField,
		ParentType: parentType,
		FullPath:   fullPath,
	}
}

func BuildAllOverlays(nodes []node.Node, allNodes []node.Node, workerCount int) error {
	nodeChan := make(chan node.Node, len(nodes))
	errChan := make(chan error, len(nodes)*2)

	var wg sync.WaitGroup
	worker := func() {
		for n := range nodeChan {
			wwlog.Info("Building system overlay image for %s", n.Id())
			wwlog.Debug("System overlays for %s: [%s]", n.Id(), strings.Join(n.SystemOverlay, ", "))
			if len(n.SystemOverlay) < 1 {
				wwlog.Warn("No system overlays defined for %s", n.Id())
			}
			if err := BuildOverlay(n, allNodes, "system", n.SystemOverlay); err != nil {
				errChan <- fmt.Errorf("could not build system overlays %v for node %s: %w", n.SystemOverlay, n.Id(), err)
			}

			wwlog.Info("Building runtime overlay image for %s", n.Id())
			wwlog.Debug("Runtime overlays for %s: [%s]", n.Id(), strings.Join(n.RuntimeOverlay, ", "))
			if len(n.RuntimeOverlay) < 1 {
				wwlog.Warn("No runtime overlays defined for %s", n.Id())
			}
			if err := BuildOverlay(n, allNodes, "runtime", n.RuntimeOverlay); err != nil {
				errChan <- fmt.Errorf("could not build runtime overlays %v for node %s: %w", n.RuntimeOverlay, n.Id(), err)
			}
		}
		wg.Done()
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}
	for _, n := range nodes {
		nodeChan <- n
	}
	close(nodeChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}
	return nil
}

func BuildSpecificOverlays(nodes []node.Node, allNodes []node.Node, overlayNames []string, workerCount int) error {
	nodeChan := make(chan node.Node, len(nodes))
	errChan := make(chan error, len(nodes))

	var wg sync.WaitGroup
	worker := func() {
		for n := range nodeChan {
			wwlog.Info("Building overlay for %s: %v", n.Id(), overlayNames)
			for _, overlayName := range overlayNames {
				err := BuildOverlay(n, allNodes, "", []string{overlayName})
				if err != nil {
					errChan <- fmt.Errorf("could not build overlay %s for node %s: %w", overlayName, n.Id(), err)
				}
			}
		}
		wg.Done()
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}
	for _, n := range nodes {
		nodeChan <- n
	}
	close(nodeChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}
	return nil
}

/*
Build overlay for the host, so no argument needs to be given
*/
func BuildHostOverlay() error {
	hostname, _ := os.Hostname()
	hostData := node.NewNode(hostname)
	wwlog.Info("Building overlay for %s: host", hostname)
	hostdir, err := Get("host")
	if err != nil {
		return err
	}
	stats, err := os.Stat(hostdir.Rootfs())
	if err != nil {
		return fmt.Errorf("could not build host overlay: %w ", err)
	}
	if !(stats.Mode() == os.FileMode(0o750|os.ModeDir) || stats.Mode() == os.FileMode(0o700|os.ModeDir)) {
		wwlog.SecWarn("Permissions of host overlay dir %s are %s (750 is considered as secure)", hostdir.Rootfs(), stats.Mode())
	}
	registry, err := node.New()
	if err != nil {
		return err
	}
	var allNodes []node.Node
	allNodes, err = registry.FindAllNodes()
	if err != nil {
		return err
	}
	return BuildOverlayIndir(hostData, allNodes, []string{"host"}, "/")
}

/*
Get all overlays present in warewulf
*/
func FindOverlays() (overlayList []string) {
	dotfilecheck, _ := regexp.Compile(`^\..*`)
	controller := config.Get()
	var files []fs.DirEntry
	if distfiles, err := os.ReadDir(controller.Paths.DistributionOverlaydir()); err != nil {
		wwlog.Warn("error reading overlays from %s: %s", controller.Paths.DistributionOverlaydir(), err)
	} else {
		files = append(files, distfiles...)
	}
	if sitefiles, err := os.ReadDir(controller.Paths.SiteOverlaydir()); err != nil {
		wwlog.Warn("error reading overlays from %s: %s", controller.Paths.SiteOverlaydir(), err)
	} else {
		files = append(files, sitefiles...)
	}
	for _, file := range files {
		wwlog.Debug("Evaluating overlay source: %s", file.Name())
		isdotfile := dotfilecheck.MatchString(file.Name())

		if file.IsDir() && !isdotfile && !util.InSlice(overlayList, file.Name()) {
			overlayList = append(overlayList, file.Name())
		}
	}
	return overlayList
}

/*
Build the given overlays for a node and create an image for them
*/
func BuildOverlay(nodeConf node.Node, allNodes []node.Node, context string, overlayNames []string) error {
	if len(overlayNames) == 0 && context == "" {
		return nil
	}

	// create the dir where the overlay images will reside
	var name string
	if context != "" {
		name = fmt.Sprintf("%s %s overlay", nodeConf.Id(), context)
	} else {
		name = fmt.Sprintf("%s overlay/%v", nodeConf.Id(), overlayNames)
	}
	overlayImage := Image(nodeConf.Id(), context, overlayNames)
	overlayImageDir := path.Dir(overlayImage)

	err := os.MkdirAll(overlayImageDir, 0o750)
	if err != nil {
		return fmt.Errorf("failed to create directory for %s: %s: %w", name, overlayImageDir, err)
	}

	wwlog.Debug("Created directory for %s: %s", name, overlayImageDir)

	buildDir, err := os.MkdirTemp(os.TempDir(), ".wwctl-overlay-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory for %s: %w", name, err)
	}
	defer os.RemoveAll(buildDir)

	wwlog.Debug("Created temporary directory for %s: %s", name, buildDir)

	err = BuildOverlayIndir(nodeConf, allNodes, overlayNames, buildDir)
	if err != nil {
		return fmt.Errorf("failed to generate files for %s: %w", name, err)
	}

	wwlog.Debug("Generated files for %s", name)

	err = util.BuildFsImage(
		name,
		buildDir,
		overlayImage,
		[]string{"*"},
		[]string{},
		// ignore cross-device files
		true,
		"newc",
		// cpio args
		"--renumber-inodes")

	return err
}

var (
	regFile *regexp.Regexp
	regLink *regexp.Regexp
)

func init() {
	regFile = regexp.MustCompile(`.*{{\s*/\*\s*file\s*["'](.*)["']\s*\*/\s*}}.*`)
	regLink = regexp.MustCompile(`.*{{\s*/\*\s*softlink\s*["'](.*)["']\s*\*/\s*}}.*`)
}

// Build the given overlays for a node in the given directory.
func BuildOverlayIndir(nodeData node.Node, allNodes []node.Node, overlayNames []string, outputDir string) error {
	if len(overlayNames) == 0 {
		return nil
	}
	if !util.IsDir(outputDir) {
		return fmt.Errorf("output must a be a directory: %s", outputDir)
	}

	if !util.ValidString(strings.Join(overlayNames, ""), "^[a-zA-Z0-9-._:]+$") {
		return fmt.Errorf("overlay names contains illegal characters: %v", overlayNames)
	}

	wwlog.Verbose("Processing node/overlays: %s/%s", nodeData.Id(), strings.Join(overlayNames, ","))
	for _, overlayName := range overlayNames {
		wwlog.Verbose("Building overlay %s for node %s in %s", overlayName, nodeData.Id(), outputDir)
		overlayRootfs, err := Get(overlayName)
		if err != nil {
			return err
		}

		wwlog.Debug("Walking the overlay structure: %s", overlayRootfs.Rootfs())
		err = filepath.Walk(overlayRootfs.Rootfs(), func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error for %s: %w", walkPath, err)
			}
			wwlog.Debug("Found overlay file: %s", walkPath)

			relPath, relErr := filepath.Rel(overlayRootfs.Rootfs(), walkPath)
			if relErr != nil {
				wwlog.Warn("Error computing relative path for %s: %v", walkPath, relErr)
				return relErr
			}
			outputPath := path.Join(outputDir, relPath)

			if info.IsDir() {
				wwlog.Debug("Found directory: %s", walkPath)

				if err = os.MkdirAll(outputPath, info.Mode()); err != nil {
					return fmt.Errorf("could not create directory within overlay: %w", err)
				}
				if err = util.CopyUIDGID(walkPath, outputPath); err != nil {
					return fmt.Errorf("failed setting permissions on overlay directory: %w", err)
				}

				wwlog.Debug("Created directory in overlay: %s", outputPath)

			} else if filepath.Ext(walkPath) == ".ww" {
				originalOutputPath := outputPath
				outputPath := strings.TrimSuffix(outputPath, ".ww")
				tstruct, err := InitStruct(overlayName, nodeData, allNodes)
				if err != nil {
					return fmt.Errorf("failed to initial data for %s: %w", nodeData.Id(), err)
				}
				tstruct.BuildSource = walkPath
				wwlog.Verbose("Evaluating overlay template file: %s", walkPath)

				buffer, backupFile, writeFile, err := RenderTemplateFile(walkPath, tstruct)
				if err != nil {
					return fmt.Errorf("failed to render template %s: %w", walkPath, err)
				}
				if !*writeFile {
					return nil
				}
				var fileBuffer bytes.Buffer
				// search for magic file name comment
				fileScanner := bufio.NewScanner(bytes.NewReader(buffer.Bytes()))
				fileScanner.Split(ScanLines)
				writingToNamedFile := false
				isLink := false
				for fileScanner.Scan() {
					line := fileScanner.Text()
					filenameFromTemplate := regFile.FindAllStringSubmatch(line, -1)
					targetFromTemplate := regLink.FindAllStringSubmatch(line, -1)
					if len(targetFromTemplate) != 0 {
						target := targetFromTemplate[0][1]
						wwlog.Debug("Creating soft link %s -> %s", outputPath, target)
						err := os.Symlink(target, outputPath)
						if err != nil {
							return fmt.Errorf("could not create symlink from template: %w", err)
						} else {
							isLink = true
						}
					} else if len(filenameFromTemplate) != 0 {
						wwlog.Debug("Writing file %s", filenameFromTemplate[0][1])
						if writingToNamedFile && !isLink {
							err = CarefulWriteBuffer(outputPath, fileBuffer, *backupFile, info.Mode())
							if err != nil {
								return fmt.Errorf("could not write file from template: %w", err)
							}
							err = util.CopyUIDGID(walkPath, outputPath)
							if err != nil {
								return fmt.Errorf("failed setting permissions on template output file: %w", err)
							}
							fileBuffer.Reset()
						}
						if path.IsAbs(filenameFromTemplate[0][1]) {
							outputPath = filenameFromTemplate[0][1]
							// Create parent directory for absolute paths
							parentDir := path.Dir(outputPath)
							sourceDirInfo, err := os.Stat(path.Dir(walkPath))
							if err != nil {
								return fmt.Errorf("could not stat source directory: %w", err)
							}
							if err := os.MkdirAll(parentDir, sourceDirInfo.Mode()); err != nil {
								return fmt.Errorf("could not create parent directory for absolute path: %w", err)
							}
						} else {
							outputPath = path.Join(path.Dir(originalOutputPath), filenameFromTemplate[0][1])
						}
						writingToNamedFile = true
						isLink = false
					} else {
						if _, err = fileBuffer.WriteString(line); err != nil {
							return fmt.Errorf("could not write to template buffer: %w", err)
						}
					}
				}
				if !isLink {
					err = CarefulWriteBuffer(outputPath, fileBuffer, *backupFile, info.Mode())
					if err != nil {
						return fmt.Errorf("could not write file from template: %w", err)
					}
					err = util.CopyUIDGID(walkPath, outputPath)
					if err != nil {
						return fmt.Errorf("failed setting permissions on template output file: %w", err)
					}
					wwlog.Debug("Wrote template file into overlay: %s", outputPath)
				}

			} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				wwlog.Debug("Found symlink %s", walkPath)
				target, err := os.Readlink(walkPath)
				if err != nil {
					return fmt.Errorf("failed reading symlink: %w", err)
				}
				if util.IsFile(outputPath) {
					backupPath := outputPath + ".wwbackup"
					if !util.IsFile(backupPath) {
						wwlog.Debug("Output file already exists: moving to backup file")
						if err = os.Rename(outputPath, backupPath); err != nil {
							return fmt.Errorf("failed renaming to backup file: %w", err)
						}
					} else {
						wwlog.Debug("%s exists, keeping the backup file", backupPath)
						if err = os.Remove(outputPath); err != nil {
							return fmt.Errorf("failed removing existing file: %w", err)
						}
					}
				}
				if err = os.Symlink(target, outputPath); err != nil {
					return fmt.Errorf("failed creating symlink: %w", err)
				}
				wwlog.Debug("Created symlink file: %s", outputPath)
			} else {
				if err := util.CopyFile(walkPath, outputPath); err != nil {
					return fmt.Errorf("could not copy file into overlay: %w", err)
				}
				wwlog.Debug("Copied overlay file: %s", outputPath)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to build overlay image directory: %w", err)
		}
	}

	return nil
}

/*
Writes buffer to the destination file. If wwbackup is set a wwbackup will be created.
*/
func CarefulWriteBuffer(destFile string, buffer bytes.Buffer, backupFile bool, perm fs.FileMode) error {
	wwlog.Debug("Trying to careful write file (%d bytes): %s", buffer.Len(), destFile)
	if backupFile {
		if !util.IsFile(destFile+".wwbackup") && util.IsFile(destFile) {
			err := util.CopyFile(destFile, destFile+".wwbackup")
			if err != nil {
				return fmt.Errorf("failed to create backup: %s -> %s.wwbackup %w", destFile, destFile, err)
			}
		}
	}
	w, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("could not open new file for template %w", err)
	}
	defer w.Close()
	_, err = buffer.WriteTo(w)
	return err
}

// getTemplateFuncMap returns a template.FuncMap with all the functions
// for warewulf templates.
func getTemplateFuncMap(fileName string, data TemplateStruct) (funcMap template.FuncMap, writeFile, backupFile *bool) {
	// Build our FuncMap
	_writeFile := true
	_backupFile := true
	writeFile = &_writeFile
	backupFile = &_backupFile
	funcMap = template.FuncMap{
		"Include":      templateFileInclude,
		"IncludeFrom":  templateImageFileInclude,
		"IncludeBlock": templateFileBlock,
		"ImportLink":   importSoftlink,
		"basename":     path.Base,
		"inc":          func(i int) int { return i + 1 },
		"dec":          func(i int) int { return i - 1 },
		"file":         func(str string) string { return fmt.Sprintf("{{ /* file \"%s\" */ }}", str) },
		"softlink":     softlink,
		"readlink":     filepath.EvalSymlinks,
		"IgnitionJson": func() string {
			return createIgnitionJson(data.ThisNode)
		},
		"abort": func() string {
			wwlog.Debug("abort file called in %s", fileName)
			*writeFile = false
			return ""
		},
		"nobackup": func() string {
			wwlog.Debug("not backup for %s", fileName)
			*backupFile = false
			return ""
		},
		"UniqueField":       UniqueField,
		"SystemdEscape":     unit.UnitNameEscape,
		"SystemdEscapePath": unit.UnitNamePathEscape,
	}

	// Merge sprig.FuncMap with our FuncMap
	for key, value := range sprig.TxtFuncMap() {
		funcMap[key] = value
	}
	return funcMap, writeFile, backupFile
}

/*
Parses the template with the given filename, variables must be in data. Returns the
parsed template as bytes.Buffer, and the bool variables for backupFile and writeFile.
If something goes wrong an error is returned.
*/
func RenderTemplateFile(fileName string, data TemplateStruct) (
	buffer bytes.Buffer, backupFile, writeFile *bool,
	err error,
) {

	funcMap, writeFile, backupFile := getTemplateFuncMap(fileName, data)

	// Create the template with the merged FuncMap
	tmpl, err := template.New(path.Base(fileName)).Option("missingkey=default").Funcs(funcMap).ParseGlob(fileName)
	if err != nil {
		err = fmt.Errorf("could not parse template %s: %w", fileName, err)
		return
	}

	err = tmpl.Execute(&buffer, data)
	if err != nil {
		err = fmt.Errorf("could not execute template: %w", err)
		return
	}
	return
}

// Simple version of ScanLines, but include the line break
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// Get all the files as a string slice for a given overlay
func (overlay Overlay) GetFiles() (files []string, err error) {
	err = filepath.Walk(overlay.Rootfs(), func(path string, info fs.FileInfo, err error) error {
		if util.IsFile(path) {
			files = append(files, strings.TrimPrefix(path, overlay.Rootfs()))
		}
		return nil
	})
	return
}
