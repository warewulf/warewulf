package edit

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

const Seperator = "# DO NOT EDIT ABOVE THIS LINE, CHANGES WILL BE LOST"

func CobraRunE(cmd *cobra.Command, args []string) error {
	if !node.CanWriteConfig() {
		return fmt.Errorf("can not write to config: exiting")
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "/bin/vi"
	}
	wwlog.Debug("using editor: %s", editor)

	registry, regErr := node.New()
	if regErr != nil {
		return regErr
	}

	if len(args) == 0 {
		for nodeID := range registry.Nodes {
			args = append(args, nodeID)
		}
	} else {
		args = hostlist.Expand(args)
	}
	wwlog.Debug("node list: %v", args)

	tempFile, tempErr := os.CreateTemp(os.TempDir(), "ww4NodeEdit*.yaml")
	if tempErr != nil {
		return fmt.Errorf("could not create temp file: %s", tempErr)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	if !NoHeader {
		yamlTemplate := node.ConfToYaml(node.Node{}, nil)
		if _, err := tempFile.WriteString("#nodename:\n#  " + strings.Join(yamlTemplate, "\n#  ") + "\n" + Seperator + "\n"); err != nil {
			return err
		}
	}

	origNodes := make(map[string]*node.Node)
	for _, nodeID := range args {
		if n, ok := registry.Nodes[nodeID]; ok {
			origNodes[nodeID] = n
		}
	}

	comments := registry.GetComments()
	strippedComments := make(yaml.CommentMap)
	for k, v := range comments {
		if strings.HasPrefix(k, "$.nodes.") {
			newKey := "$." + strings.TrimPrefix(k, "$.nodes.")
			strippedComments[newKey] = v
		}
	}

	if origYaml, err := yaml.MarshalWithOptions(origNodes, yaml.Indent(2), yaml.IndentSequence(true), yaml.WithComment(strippedComments)); err != nil {
		return err
	} else if _, err := tempFile.Write(origYaml); err != nil {
		return err
	}

	initialFile, err := os.Open(tempFile.Name())
	if err != nil {
		return err
	}
	sum1, sumErr := util.HashFile(initialFile)
	_ = initialFile.Close()
	if sumErr != nil {
		return sumErr
	}
	wwlog.Debug("original hash: %s", sum1)

	for {
		if err := util.ExecInteractive(editor, tempFile.Name()); err != nil {
			return fmt.Errorf("editor process exited with non-zero code: %w", err)
		}

		finalFile, err := os.Open(tempFile.Name())
		if err != nil {
			return err
		}
		sum2, sumErr := util.HashFile(finalFile)
		_ = finalFile.Close()
		if sumErr != nil {
			return sumErr
		}
		wwlog.Debug("edited hash: %s", sum2)

		if sum1 != sum2 {
			wwlog.Debug("modified")

			editYamlBytes, err := os.ReadFile(tempFile.Name())
			if err != nil {
				return err
			}
			editYaml := string(editYamlBytes)
			if idx := strings.Index(editYaml, Seperator); idx != -1 {
				editYaml = editYaml[idx+len(Seperator):]
			}

			editNodes := make(map[string]*node.Node)
			newCommentMap := make(yaml.CommentMap)
			if err := yaml.UnmarshalWithOptions([]byte(editYaml), &editNodes, yaml.CommentToMap(newCommentMap)); err != nil {
				wwlog.Error("%v\n", err)
				if util.Confirm("Parse error: retry") {
					continue
				} else {
					break
				}
			}

			restoredComments := make(yaml.CommentMap)
			for k, v := range newCommentMap {
				if strings.HasPrefix(k, "$.") {
					newKey := "$.nodes." + strings.TrimPrefix(k, "$.")
					restoredComments[newKey] = v
				}
			}
			registry.AddComments(restoredComments)

			var added, deleted, updated int
			for nodeID := range origNodes {
				if editNode, ok := editNodes[nodeID]; !ok || editNode == nil {
					wwlog.Verbose("delete node: %s", nodeID)
					delete(registry.Nodes, nodeID)
					deleted += 1
				}
			}
			for nodeID := range editNodes {
				if _, ok := origNodes[nodeID]; !ok {
					wwlog.Verbose("add node: %s", nodeID)
					added += 1
					registry.Nodes[nodeID] = editNodes[nodeID]
				} else if equalYaml, err := util.EqualYaml(origNodes[nodeID], editNodes[nodeID]); err != nil {
					return err
				} else if !equalYaml {
					wwlog.Verbose("update node: %s", nodeID)
					updated += 1
					registry.Nodes[nodeID] = editNodes[nodeID]
				}
			}

			if Yes || util.Confirm(fmt.Sprintf("Are you sure you want to add %d, delete %d, and update %d nodes", added, deleted, updated)) {
				if err := registry.Persist(); err != nil {
					return err
				}

				if err := warewulfd.DaemonReload(); err != nil {
					return fmt.Errorf("failed to reload warewulf daemon: %w", err)
				}
			}
			break
		} else {
			wwlog.Verbose("No changes")
			break
		}
	}

	return nil
}
