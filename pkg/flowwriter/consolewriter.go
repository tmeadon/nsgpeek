package flowwriter

import (
	"fmt"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
)

type ConsoleWriter struct {
	w          io.Writer
	table      *tablewriter.Table
	flowTuples []flowTuple
}

func NewConsoleWriter(w io.Writer) *ConsoleWriter {
	cw := ConsoleWriter{
		w: w,
	}
	cw.initTableWriter()
	return &cw
}

func (cw *ConsoleWriter) initTableWriter() {
	cw.table = tablewriter.NewWriter(cw.w)
	cw.table.SetColumnSeparator("")
	cw.table.SetRowSeparator("")
	cw.table.SetBorder(false)
	cw.table.SetTablePadding("\t")
	cw.table.SetHeaderLine(false)
	cw.table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	cw.table.SetAutoFormatHeaders(false)
	cw.table.SetHeader([]string{"time", "rule", "src_addr", "src_port", "dst_addr", "dst_port", "direction", "decision", "state", "src_to_dst_bytes", "dst_to_src_bytes"})
	cw.table.Append([]string{"", "", "", "", "", "", "", "", "", "", ""})
}

func (cw *ConsoleWriter) WriteFlowBlock(data []byte) error {
	fb, err := newFlowLogBlock(data)
	if err != nil {
		return fmt.Errorf("unable to decode flow log block: %w \n%v", err, string(data))
	}

	cw.saveFlowBlock(fb)
	return nil
}

func (cw *ConsoleWriter) saveFlowBlock(fb *flowLogBlock) {
	tuples := getFlowTuples(fb)
	cw.flowTuples = append(cw.flowTuples, tuples...)
}

func getFlowTuples(fb *flowLogBlock) (tuples []flowTuple) {
	for _, flowGroup := range fb.Properties.Flows {
		for _, flow := range flowGroup.Flows {
			for _, t := range flow.FlowTuples {
				t.Rule = flowGroup.Rule
				tuples = append(tuples, t)
			}
		}
	}

	return
}

func (cw *ConsoleWriter) Flush() {
	sortFlowTuples(cw.flowTuples)

	for _, t := range cw.flowTuples {
		cw.table.Append([]string{t.Time.Format(time.StampMilli), t.Rule, t.SourceAddress, t.SourcePort,
			t.DestAddress, t.DestPort, t.Direction, t.Decision, t.State, t.SrcToDestBytes, t.DestToSrcBytes})
	}

	fmt.Print("\n")
	cw.table.Render()
	fmt.Print("\n")
	cw.initTableWriter()
}
