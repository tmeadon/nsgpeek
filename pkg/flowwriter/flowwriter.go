package flowwriter

type FlowWriter interface {
	WriteFlowBlock(data []byte) error
	Flush()
	AddFilter(f filter)
}

type filter interface {
	Print(t flowTuple) bool
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
