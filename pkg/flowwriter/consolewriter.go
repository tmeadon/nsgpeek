package flowwriter

import (
	"fmt"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
)

type ConsoleWriter struct {
	w     io.Writer
	table *tablewriter.Table
}

func NewConsoleWriter(w io.Writer) *ConsoleWriter {
	fw := ConsoleWriter{
		w: w,
	}
	fw.initTableWriter()
	return &fw
}

func (fw *ConsoleWriter) initTableWriter() {
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

func (fw *ConsoleWriter) WriteFlowBlock(data []byte) error {
	fb, err := newFlowLogBlock(data)
	if err != nil {
		return fmt.Errorf("unable to decode flow log block: %w \n%v", err, string(data))
	}

	fw.writeFlowBlock(fb)
	return nil
}

func (fw *ConsoleWriter) writeFlowBlock(fb *flowLogBlock) {
	for _, flowGroup := range fb.Properties.Flows {
		for _, flow := range flowGroup.Flows {
			for _, t := range flow.FlowTuples {
				fw.table.Append([]string{t.Time.Format(time.StampMilli), flowGroup.Rule, t.SourceAddress, t.SourcePort,
					t.DestAddress, t.DestPort, t.Direction, t.Decision, t.State, t.SrcToDestBytes, t.DestToSrcBytes})
			}
		}
	}
}

func (fw *ConsoleWriter) Flush() {
	fmt.Print("\n")
	fw.table.Render()
	fmt.Print("\n")
	fw.initTableWriter()
}
