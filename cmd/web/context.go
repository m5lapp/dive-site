package main

import (
	"context"
	"net/http"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")
const userContextKey = contextKey("user")

// The contextSetIsAuthenticated method returns a new copy of the request with
// the provided isAuthenticated value added to the context.
func (app *app) contextSetIsAuthenticated(r *http.Request, isAuthenticated bool) *http.Request {
	ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, isAuthenticated)
	return r.WithContext(ctx)
}

// The contextGetIsAuthenticated retrieves the isAuthenticated value from the
// request context. The only time that we'll use this helper is when we
// logically expect there to be a value in the context, and if it doesn't exist
// it will firmly be an 'unexpected' error.
func (app *app) contextGetIsAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		panic("missing isAuthenticated value in request config")
	}

	return isAuthenticated
}

// The contextSetUser method returns a new copy of the request with the
// provided models.User struct added to the context.
func (app *app) contextSetUser(r *http.Request, user *models.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// The contextGetUser retrieves the User struct from the request context. The
// only time that we'll use this helper is when we logically expect there to be
// User struct value in the context, and if it doesn't exist it will firmly be
// an 'unexpected' error.
func (app *app) contextGetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		panic("missing user value in request config")
	}

	return user
}
