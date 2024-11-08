package edit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	apinode "github.com/warewulf/warewulf/internal/pkg/api/node"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	apiutil "github.com/warewulf/warewulf/internal/pkg/api/util"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v3"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	canWrite, err := apiutil.CanWriteConfig()
	if err != nil {
		return fmt.Errorf("while checking whether can write config, err: %w", err)
	}
	if !canWrite.CanWriteConfig {
		return fmt.Errorf("can not write to config exiting")
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "/bin/vi"
	}
	if len(args) == 0 {
		args = append(args, ".*")
	}
	filterList := wwapiv1.NodeList{
		Output: args,
	}
	nodeListMsg := apinode.FilteredNodes(&filterList)
	nodeMap := make(map[string]*node.NodeConf)
	// got proper yaml back
	_ = yaml.Unmarshal([]byte(nodeListMsg.NodeConfMapYaml), nodeMap)
	file, err := os.CreateTemp(os.TempDir(), "ww4NodeEdit*.yaml")
	if err != nil {
		return fmt.Errorf("could not create temp file: %s", err)
	}
	defer os.Remove(file.Name())
	for {
		_ = file.Truncate(0)
		_, _ = file.Seek(0, 0)
		if !NoHeader {
			yamlTemplate := node.UnmarshalConf(node.NodeConf{}, []string{"tagsdel"})
			_, _ = file.WriteString("#nodename:\n#  " + strings.Join(yamlTemplate, "\n#  ") + "\n")
		}
		_, _ = file.WriteString(nodeListMsg.NodeConfMapYaml)
		_, _ = file.Seek(0, 0)
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			wwlog.Error("Problems getting checksum of file %s\n", err)
		}
		sum1 := hex.EncodeToString(hasher.Sum(nil))
		err = util.ExecInteractive(editor, file.Name())
		if err != nil {
			return fmt.Errorf("editor process existed with non-zero")
		}
		_, _ = file.Seek(0, 0)
		hasher.Reset()
		if _, err := io.Copy(hasher, file); err != nil {
			wwlog.Error("Problems getting checksum of file %s\n", err)
		}
		sum2 := hex.EncodeToString(hasher.Sum(nil))
		wwlog.Debug("Hashes are before %s and after %s\n", sum1, sum2)
		if sum1 != sum2 {
			wwlog.Debug("Nodes were modified")
			modifiedNodeMap := make(map[string]*node.NodeConf)
			_, _ = file.Seek(0, 0)
			// ignore error as only may occurs under strange circumstances
			buffer, _ := io.ReadAll(file)
			err = yaml.Unmarshal(buffer, modifiedNodeMap)
			if err != nil {
				yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Got following error on parsing: %s, Retry", err))
				if yes {
					continue
				} else {
					break
				}
			}
			var checkErrors []error
			for nodeName, node := range modifiedNodeMap {
				err = node.Check()
				if err != nil {
					checkErrors = append(checkErrors, fmt.Errorf("node: %s parse error: %s", nodeName, err))
				}
			}
			if len(checkErrors) != 0 {
				yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Got following error on parsing: %s, Retry", checkErrors))
				if yes {
					continue
				} else {
					break
				}
			}

			nodeList := make([]string, len(nodeMap))
			i := 0
			for key := range nodeMap {
				nodeList[i] = key
				i++
			}
			yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d nodes", len(modifiedNodeMap)))
			if yes {
				err = apinode.NodeDelete(&wwapiv1.NodeDeleteParameter{NodeNames: nodeList, Force: true})
				if err != nil {
					wwlog.Verbose("Problem deleting nodes before modification %s", err)
				}
				buffer, _ = yaml.Marshal(modifiedNodeMap)
				newHash := apinode.Hash()
				err = apinode.NodeAddFromYaml(&wwapiv1.NodeYaml{
					NodeConfMapYaml: string(buffer),
					Hash:            newHash.Hash,
				})
				if err != nil {
					return fmt.Errorf("got following problem when writing back yaml: %s", err)
				}
				break
			}
		} else {
			break
		}
	}

	err = warewulfd.DaemonReload()
	if err != nil {
		return fmt.Errorf("failed to reload warewulf daemon: %w", err)
	}

	return nil
}
