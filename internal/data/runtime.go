package data

import (
	"fmt"
	"strconv"
)

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
