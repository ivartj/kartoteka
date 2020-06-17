package core

var (
	ErrNotFound = &Error{"Not found"}
)

type Error struct {
	message string
}

func (err *Error) Error() string {
	return err.message
}

