package print

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	conf := warewulfconf.Get()

	nb, err := yaml.MarshalWithOptions(conf, yaml.Indent(4), yaml.IndentSequence(true))
	if err != nil {
		return
	}

	fmt.Println(string(nb))
	return
}
