package status

type Step int

const (
	Idle Step = iota
	Validating
	Fetching
	Planning
	Confirming
	Executing
	Done
	Error
)

func (s Step) String() string {
	return [...]string{
		"Idle",
		"Validating",
		"Fetching",
		"Planning",
		"Confirming",
		"Executing",
		"Done",
		"Error",
	}[s]
}

type Event interface {
	isEvent()
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

type ConfirmMsg struct {
	Summary string
}

func (PhaseMsg) isEvent()    {}
func (LogMsg) isEvent()      {}
func (ConfirmMsg) isEvent()  {}
func (ProgressMsg) isEvent() {}
func (ErrMsg) isEvent()      {}

type Reporter interface {
	Report(Event)
	Confirm(summary string) bool
}
