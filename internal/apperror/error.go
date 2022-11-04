package apperror

import "encoding/json"

var (
	ErrNotFound   = NewAppError(nil, "not found", "BS-000001")
	ErrBadRequest = NewAppError(nil, "bad request", "BS-000002")
)

type AppError struct {
	Err     error  `json:"-"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func (e *AppError) Marshal() []byte {
	marshal, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return marshal
}

func NewAppError(err error, message, code string) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}
