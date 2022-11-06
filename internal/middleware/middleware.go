package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"user-balance-service/internal/apperror"
)

type errMessage struct {
	Message string `json:"messagee"`
	Code    string `json:"code"`
}

type appHandler func(w http.ResponseWriter, r *http.Request) error

func Middleware(h appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var appError *apperror.AppError
		err := h(w, r)
		if err != nil {
			if errors.As(err, &appError) {
				if errors.Is(err, apperror.ErrNotFound) {
					w.WriteHeader(404)
					w.Write(apperror.ErrNotFound.Marshal())
					return
				}
				if errors.Is(err, apperror.ErrBadRequest) {
					w.WriteHeader(400)
					w.Write(apperror.ErrBadRequest.Marshal())
					return
				}
			}

			w.WriteHeader(500)
			json.NewEncoder(w).Encode(errMessage{
				Message: "Internal error",
				Code:    "BS-000000",
			})
		}
	}
}
