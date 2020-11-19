package frazer

type Error struct {
	code int
	s string
}

func (e Error) Error() string {
	return e.s
}

func (e Error) Code() int {
	return e.code
}

func NewError(code int, message string) Error {
	return Error{
		code: code,
		s:    message,
	}
}
