package helper

import (
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type PrintHelper struct {
	*tablewriter.Table
	sb *strings.Builder
}

func New(header []string) *PrintHelper {
	sb := &strings.Builder{}
	tb := tablewriter.NewWriter(sb)
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
		sb:    sb,
	}
}

func (p *PrintHelper) String() string {
	exp := regexp.MustCompile("(?m) *$")
	return string(exp.ReplaceAll([]byte(p.sb.String()), []byte("")))
}
