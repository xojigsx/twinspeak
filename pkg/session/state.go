package session

type State int

const (
	StateConnecting State = iota
	StateConfigured
	StateActive
	StateClosing
	StateClosed
)

func (s State) String() string {
	switch s {
	case StateConnecting:
		return "Connecting"
	case StateConfigured:
		return "Configured"
	case StateActive:
		return "Active"
	case StateClosing:
		return "Closing"
	case StateClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}
