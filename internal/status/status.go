package status

type SyncStatus int

const (
	StatusIdle SyncStatus = iota
	StatusValidating
	StatusFetching
	StatusPlanning
	StatusExecuting
	StatusDone
	StatusError
)

type ProgressMsg struct {
	Status  SyncStatus
	Stage   string
	Current int
	Total   int
	Log     string
	Error   error
}

type Reporter interface {
	Report(ProgressMsg)
	Confirm(summary string) bool
}
