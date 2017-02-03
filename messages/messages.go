package messages

// Action represents commands that are sent to both display and pilot UIs
type Action struct {
	Command string
	Content string
}
