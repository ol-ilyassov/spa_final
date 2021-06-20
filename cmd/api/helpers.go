package main

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

// Retrieve "id" URL parameter from request context
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// ParamsFromContext() retrieves a slice containing parameter names and values.
	params := httprouter.ParamsFromContext(r.Context())

	// Value returned by ByName() is always a string
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}
