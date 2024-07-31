package edit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	apinode "github.com/warewulf/warewulf/internal/pkg/api/node"

	"github.com/spf13/cobra"
	apiprofile "github.com/warewulf/warewulf/internal/pkg/api/profile"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	apiutil "github.com/warewulf/warewulf/internal/pkg/api/util"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	canWrite, err := apiutil.CanWriteConfig()
	if err != nil {
		wwlog.Error("While checking whether can write config, err: %w", err)
		os.Exit(1)
	}
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
	yamlTemplate := node.UnmarshalConf(node.NodeConf{}, []string{"tagsdel"})

	// use memory file
	file, err := os.CreateTemp("/dev/shm", "ww4ProfileEdit*.yaml")
	if err != nil {
		wwlog.Error("Could not create temp file:%s \n", err)
	}
	defer func() {
		if file != nil {
			file.Close()
		}
		os.Remove(file.Name())
	}()

	hasher := sha256.New()
	var buffer bytes.Buffer
	mp := &util.MMap{}

	for {
		// reset everything
		buffer.Reset()
		hasher.Reset()

		if !NoHeader {
			_, _ = buffer.WriteString("#profilename:\n#  " + strings.Join(yamlTemplate, "\n#  ") + "\n")
		}
		_, _ = buffer.WriteString(profileListMsg.NodeConfMapYaml)
		_, err = hasher.Write(buffer.Bytes())
		if err != nil {
			return fmt.Errorf("unable to write data into hasher, err: %s", err)
		}
		sum1 := hex.EncodeToString(hasher.Sum(nil))

		// mmap memory data into file
		d, err := mp.MapToFile(buffer.Bytes(), file)
		if err != nil {
			return fmt.Errorf("unable to mmap data to the file, err: %s", err)
		}
		// data is useless here, we will map file to new data later
		_ = mp.Unmap(d)

		err = util.ExecInteractive(editor, file.Name())
		if err != nil {
			return fmt.Errorf("%s editor process exits with err: %s", editor, err)
		}

		// after edit, do remap to memory
		data, err := mp.MapFromFile(file)
		if err != nil {
			return fmt.Errorf("failed to mmap data from file, err: %s", err)
		}

		// reset hasher for new calculation
		hasher.Reset()
		hasher.Write(data)
		sum2 := hex.EncodeToString(hasher.Sum(nil))
		wwlog.Debug("Hashes are before %s and after %s\n", sum1, sum2)
		if sum1 != sum2 {
			wwlog.Debug("Profiles were modified")
			modifiedProfileMap := make(map[string]*node.NodeConf)
			// ignore error as only may occurs under strange circumstances
			err = yaml.Unmarshal(data, modifiedProfileMap)
			// after data unmarshed, we do not need the mmap data anymore
			_ = mp.Unmap(data)
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
					wwlog.Verbose("Problem deleting nodes before modification %s", err)
				}
				mdata, _ := yaml.Marshal(modifiedProfileMap)
				newHash := apinode.Hash()
				err = apiprofile.ProfileAddFromYaml(&wwapiv1.NodeAddParameter{
					NodeConfYaml: string(mdata),
					Hash:         newHash.Hash,
				})
				if err != nil {
					wwlog.Error("Got following problem when writing back yaml: %s", err)
					os.Exit(1)
				}
				break
			}
		} else {
			// as the hash value is not changed
			_ = mp.Unmap(data)
			break
		}
	}

	return nil
}
