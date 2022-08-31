package flowwriter

import (
	"fmt"
	"io"
	"time"
)

type CsvFileWriter struct {
	w          io.Writer
	flowTuples []flowTuple
	filter     filter
}

func NewCsvFileWriter(w io.Writer) (*CsvFileWriter, error) {
	c := CsvFileWriter{
		w: w,
	}
	err := c.writeHeaders()
	return &c, err
}

func (c *CsvFileWriter) AddFilter(f filter) {
	c.filter = f
}

func (c *CsvFileWriter) writeHeaders() error {
	headers := "time,rule,src_addr,src_port,dst_addr,dst_port,direction,decision,state,src_to_dst_bytes,dst_to_src_bytes"

	err := c.writeLine(headers)
	if err != nil {
		return fmt.Errorf("failed to write csv headers: %w", err)
	}

	return nil
}

func (c *CsvFileWriter) writeLine(line string) error {
	l := fmt.Sprintf("%v\n", line)
	_, err := c.w.Write([]byte(l))
	return err
}

func (c *CsvFileWriter) WriteFlowBlock(data []byte) error {
	fb, err := newFlowLogBlock(data)
	if err != nil {
		return fmt.Errorf("unable to decode flow log block: %w \n%v", err, string(data))
	}

	c.saveFlowBlock(fb)
	return nil
}

func (c *CsvFileWriter) saveFlowBlock(fb *flowLogBlock) {
	tuples := getFlowTuples(fb)

	for _, t := range tuples {
		if c.filter == nil || c.filter.Print(t) {
			c.flowTuples = append(c.flowTuples, t)
		}
	}
}

func (c *CsvFileWriter) Flush() {
	sortFlowTuples(c.flowTuples)

	for _, t := range c.flowTuples {
		line := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v", t.Time.Format(time.StampMilli), t.Rule, t.SourceAddress, t.SourcePort,
			t.DestAddress, t.DestPort, t.Direction, t.Decision, t.State, t.SrcToDestBytes, t.DestToSrcBytes)
		c.writeLine(line)
	}
}
