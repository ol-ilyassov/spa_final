package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Limit the size of request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Retrieve error on unknown fields (instead of ignoring)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		// If there is an error during decoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		// To set readably error message
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
			// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
			// for syntax errors in the JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
			// If the JSON contains a field which cannot be mapped to the target destination
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
			// If the request body exceeds 1MB in size the decode will fail
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
			// A json.InvalidUnmarshalError error will be returned if we pass a non-nil
			// pointer to Decode().
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	// Retrieve error, when JSON value is not single
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
