package flowwriter

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

type FlowWriter struct {
	tw *tabwriter.Writer
}

func NewFlowWriter(w io.Writer) *FlowWriter {
	tw := tabwriter.NewWriter(w, 0, 0, 4, ' ', tabwriter.TabIndent)
	fw := FlowWriter{tw}
	fw.writeHeader()
	return &fw
}

func (fw *FlowWriter) writeHeader() {
	fmt.Fprintf(fw.tw, "\n%v\t%v\t%v\t%v\t%v\t%v\t\n", "time", "rule", "src_addr", "src_port", "dst_addr", "dst_port")
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
				fmt.Fprintf(fw.tw, "%v\t%v\t%v\t%v\t%v\t%v", fb.Time.Format(time.RFC3339Nano), flowGroup.Rule, t.SourceAddress, t.SourcePort, t.DestAddress, t.DestPort)
			}
		}
	}
}

func (fw *FlowWriter) Flush() {
	fw.tw.Flush()
}
