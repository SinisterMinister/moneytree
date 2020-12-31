package pair

type Status string

var (
	New      Status = "NEW"
	Open     Status = "OPEN"
	Success  Status = "SUCCESS"
	Failed   Status = "FAILED"
	Canceled Status = "CANCELED"
	Broken   Status = "BROKEN"
	Reversed Status = "REVERSED"
)
