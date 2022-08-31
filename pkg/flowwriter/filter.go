package flowwriter

import "time"

type TimeFilter struct {
	Start time.Time
	End   time.Time
}

func NewTimeFilter(start time.Time, end time.Time) *TimeFilter {
	return &TimeFilter{
		Start: start,
		End:   end,
	}
}

func (f *TimeFilter) Print(t flowTuple) bool {
	return (t.Time.Equal(f.Start) || t.Time.After(f.Start)) && (t.Time.Equal(f.End) || t.Time.Before(f.End))
}
