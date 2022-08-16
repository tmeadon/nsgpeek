package flowwriter

type WriterGroup struct {
	writers []FlowWriter
}

func NewWriterGroup(w ...FlowWriter) *WriterGroup {
	return &WriterGroup{
		writers: w,
	}
}

func (wg *WriterGroup) WriteFlowBlock(data []byte) error {
	for _, w := range wg.writers {
		err := w.WriteFlowBlock(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (wg *WriterGroup) Flush() {
	for _, w := range wg.writers {
		w.Flush()
	}
}
