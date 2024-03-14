package web

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type validator interface {
	Validate() error
}

// Param returns a parameter value from the request.
func Param(r *http.Request, key string) string {
	s := chi.URLParamFromCtx(r.Context(), key)

	return s
}

// Decode reads the body of an HTTP request looking for a JSON document. The
// body is decoded into the provided value.
// If the provided value is a struct then it is checked for validation tags.
// If the value implements a validate function, it is executed.
func Decode(r *http.Request, val any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(val); err != nil {
		return err
	}

	if v, ok := val.(validator); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}
