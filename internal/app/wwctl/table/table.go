package table

import (
	"io"
	"text/tabwriter"

	"github.com/cheynewallace/tabby"
)

func Prep(parts []string) []interface{} {
	args := make([]interface{}, len(parts))
	for i, v := range parts {
		if v == "" {
			args[i] = "--"
		} else {
			args[i] = v
		}
	}
	return args
}

func New(writer io.Writer) *tabby.Tabby {
	return tabby.NewCustom(tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0))
}
