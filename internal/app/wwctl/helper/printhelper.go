package helper

import (
	"github.com/olekukonko/tablewriter"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type PrintHelper struct {
	*tablewriter.Table
}

func NewPrintHelper(header []string) *PrintHelper {
	tb := tablewriter.NewWriter(wwlog.GetLogWriterInfo())
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
