package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Error for UnmarshalJSON()
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

// This should return the JSON-encoded value for the movie runtime.
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate a string with format.
	jsonValue := fmt.Sprintf("%d mins", r)
	// Use the strconv.Quote() function on the string to wrap it in double quotes.
	// Makes valid "JSON string".
	quotedJSONValue := strconv.Quote(jsonValue)
	// Convert the quoted string value to a byte slice and return it.
	return []byte(quotedJSONValue), nil
}

// Implement a UnmarshalJSON() method on the Runtime type so that it satisfies the
// json.Unmarshaler interface. *POINTER*
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// First, unquote string
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	// Split and retrieve number from string
	parts := strings.Split(unquotedJSONValue, " ")
	// Check for initial format
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}
	// Convert to int32.
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	// Convert the int32 to a Runtime type and assign this to the receiver
	*r = Runtime(i)
	return nil
}
