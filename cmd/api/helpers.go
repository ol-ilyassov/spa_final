package main

import (
	"encoding/json"
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

type envelope map[string]interface{}

// This helper sends responses.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Ma Add whitespaces to the encoded JSON, for readability in console.
	//js, err := json.MarshalIndent(data, "", "\t")
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')
	// Now, It's safe to add any headers that we want to include.
	// We loop through the header map and add each header to the http.ResponseWriter header map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}
