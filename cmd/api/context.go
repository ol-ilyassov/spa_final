package main

import (
	"context"
	"github.com/ol-ilyassov/spa_final/internal/data"
	"net/http"
)

type contextKey string

// This constant will be used as the key for
// getting and setting user information in the request context.
const userContextKey = contextKey("user")

// Returns a new copy of the request with the provided User struct added to the context.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// Retrieves the User struct from the request context.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
