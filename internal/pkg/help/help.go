package help

// UsageTemplate replaces the default usage template from cobra
const UsageTemplate string = `Usage:
{{if .Runnable}}  {{.UseLine}}
{{end}}{{if .HasAvailableSubCommands}}  {{.CommandPath}} COMMAND [OPTIONS]
{{end}}{{if gt (len .Aliases) 0}}
Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}
Examples:
  {{.Example}}
{{end}}{{if .HasAvailableSubCommands}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}
{{end}}{{if .HasAvailableLocalFlags}}
Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasAvailableInheritedFlags}}
Global Options:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}
{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} COMMAND --help" for more information about a COMMAND.
{{end}} 
`
// End UsageTemplate

// HelpTemplate replaces the default help template from cobra
const HelpTemplate string = `
{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
