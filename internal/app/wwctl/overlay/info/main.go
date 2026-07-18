package info

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"golang.org/x/exp/maps"
)

type variableInfo struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Dynamic bool   `json:"dynamic"`
	Option  string `json:"option,omitempty"`
	Type    string `json:"type,omitempty"`
	Help    string `json:"help,omitempty"`
}

type infoOutput struct {
	Overlay  string         `json:"overlay"`
	Template string         `json:"template"`
	Node     string         `json:"node,omitempty"`
	Vars     []variableInfo `json:"variables"`
	Rendered string         `json:"rendered,omitempty"`
	Writes   *bool          `json:"writes,omitempty"`
}

func CobraRunE(cmd *cobra.Command, args []string) error {
	format := strings.ToLower(Format)
	if format != "table" && format != "json" {
		return fmt.Errorf("invalid format %q: expected table or json", Format)
	}
	if Render && strings.TrimSpace(NodeName) == "" {
		return fmt.Errorf("--render requires --node")
	}

	overlayName := args[0]
	filePath := args[1]

	ov, err := overlay.Get(overlayName)
	if err != nil {
		wwlog.Error("Failed to get overlay %s: %s", overlayName, err)
		return err
	}

	varFields := ov.ParseVarFields(filePath)
	if varFields == nil {
		return fmt.Errorf("could not parse variables for %s in overlay %s", filePath, overlayName)
	}
	commentMap := ov.ParseCommentVars(filePath)

	if strings.TrimSpace(NodeName) != "" {
		return runNodeInfo(cmd, ov, overlayName, filePath, varFields, commentMap, format)
	}
	if format == "json" {
		return printStaticJSON(cmd, overlayName, filePath, varFields, commentMap)
	}

	return printStaticTable(cmd, varFields, commentMap)
}

func printStaticTable(cmd *cobra.Command, varFields map[string]overlay.FieldInfo, commentMap map[string]string) error {
	commentKeys := maps.Keys(commentMap)
	sort.Strings(commentKeys)
	hasWwdoc := false
	for _, docLn := range commentKeys {
		if strings.Contains(docLn, "wwdoc") {
			wwlog.Info(commentMap[docLn])
			hasWwdoc = true
		}
	}

	// Add newline after wwdoc lines if they exist
	if hasWwdoc {
		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Sort variables by name for consistent output
	varNames := maps.Keys(varFields)
	sort.Strings(varNames)

	t := table.New(cmd.OutOrStdout())
	t.AddHeader("VARIABLE", "OPTION", "TYPE", "HELP")

	for _, varName := range varNames {
		fieldInfo := varFields[varName]
		helpText, hasCommentHelp := commentMap[varName]

		// Extract metadata from the resolved field
		opt := ""
		typ := ""
		help := ""

		// Check if we have valid field information (field name is not empty)
		hasValidField := fieldInfo.Field.Name != ""

		if hasValidField {
			// Get option from lopt tag
			if lopt := fieldInfo.Field.Tag.Get("lopt"); lopt != "" {
				opt = "--" + lopt
			}

			// Get type from type tag, or use field type string
			typ = fieldInfo.Field.Tag.Get("type")
			if typ == "" {
				// Use String() instead of Name() to handle composite types (slices, maps, pointers)
				typ = fieldInfo.Field.Type.String()
			}

			// Get help from comment tag
			help = fieldInfo.Field.Tag.Get("comment")
		}

		// Prefer inline comment documentation if available
		if hasCommentHelp {
			help = helpText
		}

		// Special handling for Tags fields
		if strings.Contains(varName, "Tags") {
			t.AddLine(varName, opt, typ, help)
		} else if hasValidField || hasCommentHelp {
			t.AddLine(varName, opt, typ, help)
		}
	}

	t.Print()
	return nil
}

func printStaticJSON(cmd *cobra.Command, overlayName string, filePath string, varFields map[string]overlay.FieldInfo, commentMap map[string]string) error {
	vars := buildVariableInfo(varFields, commentMap, nil)
	data, err := json.MarshalIndent(infoOutput{Overlay: overlayName, Template: filePath, Vars: vars}, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", data)
	return err
}

func runNodeInfo(cmd *cobra.Command, ov overlay.Overlay, overlayName string, filePath string, varFields map[string]overlay.FieldInfo, commentMap map[string]string, format string) error {
	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("could not open node configuration: %w", err)
	}
	nodeData, err := nodeDB.GetNode(NodeName)
	if err != nil {
		return fmt.Errorf("could not get node %s: %w", NodeName, err)
	}
	allNodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return fmt.Errorf("could not get node list: %w", err)
	}

	tstruct, err := overlay.InitStruct(overlayName, nodeData, allNodes)
	if err != nil {
		return fmt.Errorf("could not initialize template context for node %s: %w", NodeName, err)
	}
	templatePath := ov.File(filePath)
	tstruct.BuildSource = templatePath

	vars := buildVariableInfo(varFields, commentMap, &tstruct)
	out := infoOutput{Overlay: overlayName, Template: filePath, Node: NodeName, Vars: vars}
	if Render {
		buffer, _, writeFile, err := overlay.RenderTemplateFile(templatePath, tstruct)
		if err != nil {
			return fmt.Errorf("could not render template %s: %w", filePath, err)
		}
		out.Rendered = buffer.String()
		out.Writes = writeFile
	}

	if format == "json" {
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", data)
		return err
	}

	printNodeTable(cmd, vars)
	if Render {
		if out.Writes != nil && !*out.Writes {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nRendered output: <template aborted; no output written>")
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nRendered output:\n%s", out.Rendered)
			if !strings.HasSuffix(out.Rendered, "\n") {
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
			}
		}
	}
	return nil
}

func buildVariableInfo(varFields map[string]overlay.FieldInfo, commentMap map[string]string, tstruct *overlay.TemplateStruct) []variableInfo {
	varNames := maps.Keys(varFields)
	sort.Strings(varNames)
	vars := make([]variableInfo, 0, len(varNames))
	for _, varName := range varNames {
		fieldInfo := varFields[varName]
		helpText, hasCommentHelp := commentMap[varName]

		opt := ""
		typ := ""
		help := ""
		hasValidField := fieldInfo.Field.Name != ""
		if hasValidField {
			if lopt := fieldInfo.Field.Tag.Get("lopt"); lopt != "" {
				opt = "--" + lopt
			}
			typ = fieldInfo.Field.Tag.Get("type")
			if typ == "" {
				typ = fieldInfo.Field.Type.String()
			}
			help = fieldInfo.Field.Tag.Get("comment")
		}
		if hasCommentHelp {
			help = helpText
		}

		if strings.Contains(varName, "Tags") || hasValidField || hasCommentHelp {
			entry := variableInfo{Name: varName, Option: opt, Type: typ, Help: help}
			if tstruct != nil {
				value, dynamic := resolveTemplateValue(*tstruct, varName, fieldInfo)
				entry.Value = value
				entry.Dynamic = dynamic
			}
			vars = append(vars, entry)
		}
	}
	return vars
}

func printNodeTable(cmd *cobra.Command, vars []variableInfo) {
	t := table.New(cmd.OutOrStdout())
	t.AddHeader("VARIABLE", "VALUE", "OPTION", "TYPE", "HELP")
	for _, variable := range vars {
		t.AddLine(variable.Name, variable.Value, variable.Option, variable.Type, variable.Help)
	}
	t.Print()
}

func resolveTemplateValue(tstruct overlay.TemplateStruct, varName string, fieldInfo overlay.FieldInfo) (string, bool) {
	if strings.HasPrefix(varName, "$") && !strings.HasPrefix(varName, "$.") {
		if value, ok := resolveDynamicRangeValue(tstruct, varName, fieldInfo); ok {
			return value, true
		}
		return "<dynamic>", true
	}
	pathName := strings.TrimPrefix(varName, "$")
	pathName = strings.TrimPrefix(pathName, ".")
	if pathName == "" {
		return summarizeValue(reflect.ValueOf(tstruct)), false
	}

	value, ok := resolveFieldPath(reflect.ValueOf(tstruct), strings.Split(pathName, "."))
	if !ok {
		return "<unresolved>", false
	}
	return summarizeValue(value), false
}

func resolveDynamicRangeValue(tstruct overlay.TemplateStruct, varName string, fieldInfo overlay.FieldInfo) (string, bool) {
	if fieldInfo.FullPath == "" {
		return "", false
	}
	parts := strings.Split(strings.TrimPrefix(varName, "$"), ".")
	if len(parts) < 2 {
		return fmt.Sprintf("<dynamic: %s>", fieldInfo.FullPath), true
	}
	fieldParts := parts[1:]

	rangePathDisplay := fieldInfo.FullPath
	suffix := "." + strings.Join(fieldParts, ".")
	if strings.HasSuffix(rangePathDisplay, suffix) {
		rangePathDisplay = strings.TrimSuffix(rangePathDisplay, suffix)
	}
	rangePath := strings.TrimPrefix(rangePathDisplay, ".")
	if rangePath == "" {
		return "", false
	}
	collection, ok := resolveFieldPath(reflect.ValueOf(tstruct), strings.Split(rangePath, "."))
	if !ok {
		return fmt.Sprintf("<dynamic: %s unresolved>", rangePathDisplay), true
	}
	collection = dereference(collection)
	if !collection.IsValid() {
		return fmt.Sprintf("<dynamic: %s is nil>", rangePathDisplay), true
	}

	switch collection.Kind() {
	case reflect.Map:
		if collection.Len() == 0 {
			return fmt.Sprintf("<dynamic: %s has 0 entries>", rangePathDisplay), true
		}
		keys := collection.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
		})
		values := make([]string, 0, len(keys))
		for _, key := range keys {
			entry := collection.MapIndex(key)
			fieldValue, ok := resolveFieldPath(entry, fieldParts)
			if !ok {
				values = append(values, fmt.Sprintf("%v=<unresolved>", key.Interface()))
				continue
			}
			values = append(values, fmt.Sprintf("%v=%s", key.Interface(), summarizeDynamicValue(fieldValue)))
		}
		return strings.Join(values, ", "), true
	case reflect.Slice, reflect.Array:
		if collection.Len() == 0 {
			return fmt.Sprintf("<dynamic: %s has 0 entries>", rangePathDisplay), true
		}
		values := make([]string, 0, collection.Len())
		for i := 0; i < collection.Len(); i++ {
			fieldValue, ok := resolveFieldPath(collection.Index(i), fieldParts)
			if !ok {
				values = append(values, fmt.Sprintf("%d=<unresolved>", i))
				continue
			}
			values = append(values, fmt.Sprintf("%d=%s", i, summarizeDynamicValue(fieldValue)))
		}
		return strings.Join(values, ", "), true
	}
	return fmt.Sprintf("<dynamic: %s is %s>", rangePathDisplay, collection.Kind()), true
}

func summarizeDynamicValue(value reflect.Value) string {
	summary := summarizeValue(value)
	if summary == "" || summary == "<nil>" {
		return "<empty>"
	}
	return summary
}

func resolveFieldPath(value reflect.Value, parts []string) (reflect.Value, bool) {
	current := value
	for _, part := range parts {
		current = dereference(current)
		if !current.IsValid() {
			return reflect.Value{}, false
		}
		switch current.Kind() {
		case reflect.Struct:
			field := current.FieldByName(part)
			if field.IsValid() {
				current = field
				continue
			}
			method := current.MethodByName(part)
			if method.IsValid() && method.Type().NumIn() == 0 && method.Type().NumOut() == 1 {
				current = method.Call(nil)[0]
				continue
			}
			return reflect.Value{}, false
		case reflect.Map:
			if current.Type().Key().Kind() != reflect.String {
				return reflect.Value{}, false
			}
			mapped := current.MapIndex(reflect.ValueOf(part))
			if !mapped.IsValid() {
				return reflect.ValueOf("<unset>"), true
			}
			current = mapped
		default:
			return reflect.Value{}, false
		}
	}
	return current, true
}

func dereference(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func summarizeValue(value reflect.Value) string {
	value = dereference(value)
	if !value.IsValid() {
		return "<nil>"
	}
	if value.CanInterface() {
		switch v := value.Interface().(type) {
		case fmt.Stringer:
			return v.String()
		}
	}
	switch value.Kind() {
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return fmt.Sprintf("%t", value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprintf("%d", value.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", value.Float())
	case reflect.Map:
		return fmt.Sprintf("%d entries", value.Len())
	case reflect.Slice, reflect.Array:
		return fmt.Sprintf("%d entries", value.Len())
	case reflect.Struct:
		return value.Type().String()
	}
	if value.CanInterface() {
		return fmt.Sprint(value.Interface())
	}
	return fmt.Sprintf("<%s>", value.Kind())
}
