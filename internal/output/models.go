package output

type Mode int

const (
	HIDDEN Mode = iota
	REPORT_FILE
	STDOUT
	BOTH
)

func Convert(s string) Mode {
	switch s {
	case "STDOUT":
		return STDOUT
	}
	return HIDDEN
}
