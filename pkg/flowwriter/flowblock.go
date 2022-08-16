package flowwriter

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type flowLogBlock struct {
	Time       time.Time              `json:"time"`
	Properties flowLogBlockProperties `json:"properties"`
}

func newFlowLogBlock(data []byte) (*flowLogBlock, error) {
	// strip the first rune if it's a comma to ensure we have valid json
	if r, s := utf8.DecodeRune(data); r == rune(',') {
		data = data[s:]
	}

	var fb flowLogBlock
	if err := json.Unmarshal(data, &fb); err != nil {
		return nil, err
	}

	return &fb, nil
}

type flowLogBlockProperties struct {
	Flows []flowLogBlockFlowGroup `json:"flows"`
}

type flowLogBlockFlowGroup struct {
	Rule  string             `json:"rule"`
	Flows []flowLogBlockFlow `json:"flows"`
}

type flowLogBlockFlow struct {
	Mac        string      `json:"mac"`
	FlowTuples []flowTuple `json:"flowTuples"`
}

func (f *flowLogBlockFlow) UnmarshalJSON(data []byte) error {
	var jf jsonFlowLogBlockFlow
	if err := json.Unmarshal(data, &jf); err != nil {
		return err
	}

	*f = *jf.flowLogBlockFlow()

	return nil
}

type flowTuple struct {
	Time           time.Time
	SourceAddress  string
	SourcePort     string
	DestAddress    string
	DestPort       string
	Direction      string
	Decision       string
	State          string
	SrcToDestBytes string
	DestToSrcBytes string
}

type jsonFlowLogBlockFlow struct {
	Mac        string   `json:"mac"`
	FlowTuples []string `json:"flowTuples"`
}

func (jf *jsonFlowLogBlockFlow) flowLogBlockFlow() *flowLogBlockFlow {
	tuples := make([]flowTuple, 0)

	for _, tuple := range jf.FlowTuples {
		t := strings.Split(tuple, ",")

		newTuple := flowTuple{
			Time:          convertTime(t[0]),
			SourceAddress: t[1],
			SourcePort:    t[3],
			DestAddress:   t[2],
			DestPort:      t[4],
			Direction:     formatDirection(t[6]),
			Decision:      formatDecision(t[7]),
		}

		// include the flow log v2 properties if present
		if len(t) > 8 {
			newTuple.State = formatState(t[8])
			newTuple.SrcToDestBytes = t[10]
			newTuple.DestToSrcBytes = t[12]
		}

		tuples = append(tuples, newTuple)
	}

	return &flowLogBlockFlow{
		Mac:        jf.Mac,
		FlowTuples: tuples,
	}
}

func formatDirection(dir string) string {
	if dir == "I" {
		return "in"
	} else {
		return "out"
	}
}

func formatDecision(dec string) string {
	if dec == "A" {
		return "allow"
	} else {
		return "deny"
	}
}

func formatState(state string) string {
	switch state {
	case "B":
		return "begin"
	case "C":
		return "continuing"
	case "E":
		return "end"
	default:
		return "-"
	}
}

func convertTime(unixTime string) time.Time {
	t, _ := strconv.Atoi(unixTime)
	return time.Unix(int64(t), 0).UTC()
}
