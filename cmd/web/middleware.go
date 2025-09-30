package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/m5lapp/divesite-monolith/internal/models"
)

// authenticate checks if a user has previosuly authenticated and therefore has
// an "authenticatedUserID" value in their session. A boolean "isAuthenticated"
// value will then be added to the request Context, along with their models.User
// struct if isAuthenticated is true.
func (app *app) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			r = app.contextSetIsAuthenticated(r, false)
			next.ServeHTTP(w, r)
			return
		}

		// If a non-zero user ID has been provided in the sessionManager, then
		// the user has authenticated previously and so a valid user account
		// should be guarenteed to be returned.
		user, err := app.users.GetByID(id)
		if err != nil && !errors.Is(err, models.ErrNoRecord) {
			app.serverError(w, r, fmt.Errorf("failed to fetch user with id %d: %w", id, err))
			return
		}

		// err will not be nil if it's a models.ErrNoRecord, i.e. the user could
		// not be found.
		isAuthenticated := err == nil
		r = app.contextSetIsAuthenticated(r, isAuthenticated)
		if isAuthenticated {
			r = app.contextSetUser(r, &user)
		}

		next.ServeHTTP(w, r)
	})
}

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com",
		)
		w.Header().Set("Server", "Go")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func (app *app) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.log.Info(
			"Request received",
			"ip", r.RemoteAddr,
			"protocol", r.Proto,
			"method", r.Method,
			"uri", r.URL.RequestURI(),
		)

		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (app *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "Close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *app) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			app.sessionManager.Put(r.Context(), "redirectPathAfterLogIn", r.URL.Path)
			http.Redirect(w, r, "/user/log-in", http.StatusSeeOther)
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}
