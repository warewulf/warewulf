package reference

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	linkHandler := func(name, ref string) string {
		return fmt.Sprintf(":ref:`%s <%s>`", name, ref)
	}
	cmd.Parent().Parent().DisableAutoGenTag = true
	err = doc.GenReSTTreeCustom(cmd.Parent().Parent(), args[0], func(arg string) string { return "" }, linkHandler)
	//err = doc.GenReSTTree(cmd.Parent().Parent(), args[0])
	return
}
