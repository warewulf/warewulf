package print

import (
	"fmt"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	conf := warewulfconf.New()
	buffer, err := yaml.Marshal(&conf)
	if err != nil {
		return
	}
	fmt.Println(string(buffer))
	return
}
