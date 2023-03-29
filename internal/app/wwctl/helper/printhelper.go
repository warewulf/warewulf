package helper

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

type PrintHelper struct {
	*tablewriter.Table
}

func NewPrintHelper(header []string) *PrintHelper {
	tb := tablewriter.NewWriter(os.Stdout)
	tb.SetHeader(header)
	tb.SetAutoWrapText(false)
	tb.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tb.SetAlignment(tablewriter.ALIGN_LEFT)
	tb.SetCenterSeparator("")
	tb.SetColumnSeparator("")
	tb.SetRowSeparator("")
	tb.SetHeaderLine(false)
	tb.SetBorder(false)
	return &PrintHelper{
		Table: tb,
	}
}
