package print

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
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
