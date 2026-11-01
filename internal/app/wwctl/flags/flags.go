package flags

import (
	"github.com/spf13/cobra"
)

// AddContainer adds the deprecated --container (-C) alias for --image.
func AddContainer(cmd *cobra.Command) {
	addAlias(cmd, "container", "C", "image")
}

// AddWwinit adds the deprecated --wwinit alias for --system-overlays.
func AddWwinit(cmd *cobra.Command) {
	addAlias(cmd, "wwinit", "", "system-overlays")
}

// AddRuntime adds the deprecated --runtime alias for --runtime-overlays.
func AddRuntime(cmd *cobra.Command) {
	addAlias(cmd, "runtime", "", "runtime-overlays")
}

// addAlias adds a hidden, deprecated flag that sets the flag named target.
// Setting the alias marks target as changed, so commands that only apply
// changed flags (e.g., via node.UpdateFrom) honor the alias.
func addAlias(cmd *cobra.Command, name, shorthand, target string) {
	cmd.Flags().VarP(&aliasValue{cmd: cmd, target: target}, name, shorthand, "alias for --"+target)
	cmd.Flags().Lookup(name).Hidden = true
	if err := cmd.Flags().MarkDeprecated(name, "use --"+target+" instead"); err != nil {
		panic(err)
	}
}

// aliasValue is the flag value of an alias. It forwards to the target flag,
// which may be registered as a persistent flag and so is only merged into
// cmd.Flags() when the command parses its arguments.
type aliasValue struct {
	cmd    *cobra.Command
	target string
}

func (a *aliasValue) Set(value string) error {
	if a.cmd.Flags().Lookup(a.target) != nil {
		return a.cmd.Flags().Set(a.target, value)
	}
	return a.cmd.PersistentFlags().Set(a.target, value)
}

func (a *aliasValue) String() string {
	if f := a.cmd.Flags().Lookup(a.target); f != nil {
		return f.Value.String()
	}
	if f := a.cmd.PersistentFlags().Lookup(a.target); f != nil {
		return f.Value.String()
	}
	return ""
}

func (a *aliasValue) Type() string {
	if f := a.cmd.Flags().Lookup(a.target); f != nil {
		return f.Value.Type()
	}
	if f := a.cmd.PersistentFlags().Lookup(a.target); f != nil {
		return f.Value.Type()
	}
	return "string"
}
