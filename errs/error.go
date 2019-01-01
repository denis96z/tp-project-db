package errs

import "net/http"

//go:generate easyjson

//easyjson:json
type Error struct {
	HttpStatus int    `json:"-"`
	Message    string `json:"message"`
}

func NewError(status int, message string) *Error {
	return &Error{
		HttpStatus: status,
		Message:    message,
	}
}

func NewInternalError(message string) *Error {
	return NewError(http.StatusInternalServerError, message)
}

func NewNotFoundError(message string) *Error {
	return NewError(http.StatusNotFound, message)
}

func NewInvalidFormatError(message string) *Error {
	return NewError(http.StatusUnprocessableEntity, message)
}

func NewConflictError(message string) *Error {
	return NewError(http.StatusConflict, message)
}

func (err Error) Error() string {
	return err.Message
}
