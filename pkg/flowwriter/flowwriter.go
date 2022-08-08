package flowwriter

import (
	"fmt"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
)

type FlowWriter struct {
	w     io.Writer
	table *tablewriter.Table
}

func NewFlowWriter(w io.Writer) *FlowWriter {
	fw := FlowWriter{
		w: w,
	}
	fw.initTableWriter()
	return &fw
}

func (fw *FlowWriter) initTableWriter() {
	fw.table = tablewriter.NewWriter(fw.w)
	fw.table.SetColumnSeparator("")
	fw.table.SetRowSeparator("")
	fw.table.SetBorder(false)
	fw.table.SetTablePadding("\t")
	fw.table.SetHeaderLine(false)
	fw.table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	fw.table.SetAutoFormatHeaders(false)
	fw.table.SetHeader([]string{"time", "rule", "src_addr", "src_port", "dst_addr", "dst_port", "direction", "decision", "state", "src_to_dst_bytes", "dst_to_src_bytes"})
	fw.table.Append([]string{"", "", "", "", "", "", "", "", "", "", ""})
}

func (fw *FlowWriter) WriteFlowBlock(data []byte) error {
	fb, err := NewFlowLogBlock(data)
	if err != nil {
		return fmt.Errorf("unable to decode flow log block: %w", err)
	}

	fw.writeFlowBlock(fb)
	return nil
}

func (fw *FlowWriter) writeFlowBlock(fb *FlowLogBlock) {
	for _, flowGroup := range fb.Properties.Flows {
		for _, flow := range flowGroup.Flows {
			for _, t := range flow.FlowTuples {
				fw.table.Append([]string{fb.Time.Format(time.StampMilli), flowGroup.Rule, t.SourceAddress, t.SourcePort,
					t.DestAddress, t.DestPort, t.Direction, t.Decision, t.State, t.SrcToDestBytes, t.DestToSrcBytes})
			}
		}
	}
}

func (fw *FlowWriter) Flush() {
	fmt.Print("\n")
	fw.table.Render()
	fmt.Print("\n")
	fw.initTableWriter()
}
