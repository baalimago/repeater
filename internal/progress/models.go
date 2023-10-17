package progress

type Mode int

const (
	HIDDEN Mode = iota
	REPORT_FILE
	STDOUT
	BOTH
)
