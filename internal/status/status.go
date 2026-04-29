package status

type Step int

const (
	Idle Step = iota
	Validating
	Fetching
	Planning
	Executing
	Done
	Error
)

type Event interface {
	syncEvent()
}

type ProgressMsg struct {
	Current int
	Total   int
}

type ErrMsg struct {
	Error error
}

type LogMsg struct {
	Text string
}

type PhaseMsg struct {
	Step  Step
	Label string
}

type PlanMsg struct {
	Summary string
}

func (PhaseMsg) syncEvent()    {}
func (LogMsg) syncEvent()      {}
func (PlanMsg) syncEvent()     {}
func (ProgressMsg) syncEvent() {}
func (ErrMsg) syncEvent()      {}

type Reporter interface {
	Report(Event)
	Confirm(summary string) bool
}
