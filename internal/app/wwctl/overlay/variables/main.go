package variables

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/node"
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

	vars := ov.ParseVars(filePath)
	sort.Strings(vars)
	commentMap := ov.ParseCommentVars(filePath)
	varMap := node.TemplateVarMap{}
	varMap.ConfToTemplateMap(node.Node{}, "")
	if vars == nil {
		return fmt.Errorf("could not parse variables for %s in overlay %s", filePath, overlayName)
	}
	commentKeys := maps.Keys(commentMap)
	sort.Strings(commentKeys)
	for _, docLn := range commentKeys {
		if strings.Contains(docLn, "wwdoc") {
			wwlog.Info(commentMap[docLn])
		}
	}
	t := table.New(cmd.OutOrStdout())
	t.AddHeader("OVERLAY VARIABLE", "HELP", "TYPE", "CMD OPTION")

	for _, v := range vars {
		found := false
		helpText, hasCommentHelp := commentMap[v]

		for key, val := range varMap {
			// fuzzy match, ignore case and try to also match singular / plural
			textLower := strings.ToLower(v)
			keyLower := strings.ToLower(key)
			match := false
			if strings.Contains(textLower, keyLower) {
				match = true
			} else {
				keyParts := strings.Split(key, ".")
				newParts := make([]string, len(keyParts))
				for i, p := range keyParts {
					newParts[i] = strings.ToLower(p)
				}
				for i, part := range newParts {
					originalPart := part
					var variation string
					if strings.HasSuffix(part, "s") {
						variation = strings.TrimSuffix(part, "s")
					} else {
						variation = part + "s"
					}
					newParts[i] = variation
					variantKey := strings.Join(newParts, ".")
					if strings.Contains(textLower, variantKey) {
						match = true
						break
					}
					newParts[i] = originalPart // restore for next iteration
				}
			}
			if match {
				opt := ""
				if val.LongOpt != "" {
					opt = val.LongOpt
				}
				if hasCommentHelp {
					t.AddLine(v, helpText, val.Type, opt)
				} else {
					t.AddLine(v, val.Comment, val.Type, opt)
				}
				found = true
			}
		}
		if !found {
			if hasCommentHelp {
				t.AddLine(v, helpText, "", "")
			} else if strings.Contains(v, "Tags") {
				t.AddLine(v, "", "", "", "")
			}
		}
	}
	t.Print()
	return nil
}
