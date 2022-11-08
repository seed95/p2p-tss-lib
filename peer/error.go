package peer

//TODO check this struct and codes
// Checked with tss.Error

type Error struct {
	Message string
	ID      string
}

func (e Error) Error() string {
	return e.Message
}
