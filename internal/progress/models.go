package progress

type Mode int

const (
	HIDDEN Mode = iota
	REPORT_FILE
	STDOUT
	BOTH
)

func New(s *string) Mode {
	switch *s {
	case "HIDDEN":
		return HIDDEN
	case "REPORT_FILE":
		return REPORT_FILE
	case "STDOUT":
		return STDOUT
	case "BOTH":
		return BOTH
	default:
		return STDOUT
	}
}
func (m Mode) String() string {
	switch m {
	case HIDDEN:
		return "HIDDEN"
	case REPORT_FILE:
		return "REPORT_FILE"
	case STDOUT:
		return "STDOUT"
	case BOTH:
		return "BOTH"
	}
	return "UNKNOWN_TYPE"
}
