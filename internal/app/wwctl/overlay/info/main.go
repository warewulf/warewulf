package info

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"golang.org/x/exp/maps"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	overlayName := args[0]
	filePath := args[1]

	ov, err := overlay.Get(overlayName)
	if err != nil {
		wwlog.Error("Failed to get overlay %s: %s", overlayName, err)
		return err
	}

	// Use new type-based variable resolution
	varFields := ov.ParseVarFields(filePath)
	if varFields == nil {
		return fmt.Errorf("could not parse variables for %s in overlay %s", filePath, overlayName)
	}

	// Still parse comment vars for wwdoc and inline documentation
	commentMap := ov.ParseCommentVars(filePath)
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
