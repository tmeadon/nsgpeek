package flowwriter

type FlowWriter interface {
	WriteFlowBlock(data []byte) error
	Flush()
}
