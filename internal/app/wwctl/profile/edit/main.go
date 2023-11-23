package edit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"

	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	apiutil "github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	canWrite := apiutil.CanWriteConfig()
	if !canWrite.CanWriteConfig {
		wwlog.Error("Can't write to config exiting")
		os.Exit(1)
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
	profileListMsg := apiprofile.FilteredProfiles(&filterList)
	profileMap := make(map[string]*node.NodeConf)
	// got proper yaml back
	_ = yaml.Unmarshal([]byte(profileListMsg.NodeConfMapYaml), profileMap)
	file, err := os.CreateTemp(os.TempDir(), "ww4ProfileEdit*.yaml")
	if err != nil {
		wwlog.Error("Could not create temp file:%s \n", err)
	}
	defer os.Remove(file.Name())
	yamlTemplate := node.UnmarshalConf(node.NodeConf{}, []string{"tagsdel"})
	for {
		_ = file.Truncate(0)
		_, _ = file.Seek(0, 0)
		if !NoHeader {
			_, _ = file.WriteString("#profilename:\n#  " + strings.Join(yamlTemplate, "\n#  ") + "\n")
		}
		_, _ = file.WriteString(profileListMsg.NodeConfMapYaml)
		_, _ = file.Seek(0, 0)
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			wwlog.Error("Problems getting checksum of file %s\n", err)
		}
		sum1 := hex.EncodeToString(hasher.Sum(nil))
		err = util.ExecInteractive(editor, file.Name())
		if err != nil {
			wwlog.Error("Editor process existed with non-zero\n")
			os.Exit(1)
		}
		_, _ = file.Seek(0, 0)
		hasher.Reset()
		if _, err := io.Copy(hasher, file); err != nil {
			wwlog.Error("Problems getting checksum of file %s\n", err)
		}
		sum2 := hex.EncodeToString(hasher.Sum(nil))
		wwlog.Debug("Hashes are before %s and after %s\n", sum1, sum2)
		if sum1 != sum2 {
			wwlog.Debug("Profiles were modified")
			modifiedProfileMap := make(map[string]*node.NodeConf)
			_, _ = file.Seek(0, 0)
			// ignore error as only may occurs under strange circumstances
			buffer, _ := io.ReadAll(file)
			err = yaml.Unmarshal(buffer, modifiedProfileMap)
			if err != nil {
				yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Got following error on parsing: %s, Retry", err))
				if yes {
					continue
				} else {
					break
				}
			}
			var checkErrors []error
			for nodeName, node := range modifiedProfileMap {
				err = node.Check()
				if err != nil {
					checkErrors = append(checkErrors, fmt.Errorf("profile: %s parse error: %s", nodeName, err))
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
			pList := make([]string, len(profileMap))
			i := 0
			for key := range profileMap {
				pList[i] = key
				i++
			}
			yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d nodes", len(modifiedProfileMap)))
			if yes {
				err = apiprofile.ProfileDelete(&wwapiv1.NodeDeleteParameter{NodeNames: pList, Force: true})

				if err != nil {
					wwlog.Verbose("Problem deleting nodes before modification %s")
				}
				buffer, _ = yaml.Marshal(modifiedProfileMap)
				newHash := apinode.Hash()
				err = apiprofile.ProfileAddFromYaml(&wwapiv1.NodeAddParameter{
					NodeConfYaml: string(buffer),
					Hash:         newHash.Hash,
				})
				if err != nil {
					wwlog.Error("Got following problem when writing back yaml: %s", err)
					os.Exit(1)
				}
				break
			}
		} else {
			break
		}
	}

	return nil
}
