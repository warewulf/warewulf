package helper

import (
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/olekukonko/tablewriter"
)

type PrintHelper struct {
	*tablewriter.Table
}

func NewPrintHelper(header []string) *PrintHelper {
	tb := tablewriter.NewWriter(wwlog.GetLogWriter())
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
