package main

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/m5lapp/divesite-monolith/internal/assert"
)

func TestStatus(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/status")

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}

func TestDiveSiteGet(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_ = ts.logIn(t, "", "")

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/log-book/dive-site/view/1",
			wantCode: http.StatusOK,
			wantBody: "Sail Rock",
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/log-book/dive-site/view/99999",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/log-book/dive-site/view/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/log-book/dive-site/view/3.14159",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/log-book/dive-site/view/foo",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty ID",
			urlPath:  "/log-book/dive-site/view/",
			wantCode: http.StatusOK,
			wantBody: "Sail Rock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestUserSignUp(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/user/sign-up")
	csrfToken := extractCSRFToken(t, body)

	const (
		validName     = "Bob"
		validPassword = "Pa55W0rd"
		validEmail    = "bob@example.com"
		formTag       = `<form method="post" action="/user/sign-up"`
	)

	tests := []struct {
		name                string
		userName            string
		userEmail           string
		userPassword        string
		userPasswordConfirm string
		csrfToken           string
		wantCode            int
		wantFormTag         string
	}{
		{
			name:                "Valid submission",
			userName:            validName,
			userEmail:           validEmail,
			userPassword:        validPassword,
			userPasswordConfirm: validPassword,
			csrfToken:           csrfToken,
			wantCode:            http.StatusSeeOther,
		},
		{
			name:                "Invalid CSRF token",
			userName:            validName,
			userEmail:           validEmail,
			userPassword:        validPassword,
			userPasswordConfirm: validPassword,
			csrfToken:           "invalid_token",
			wantCode:            http.StatusBadRequest,
		},
		{
			name:                "Empty name",
			userName:            "",
			userEmail:           validEmail,
			userPassword:        validPassword,
			userPasswordConfirm: validPassword,
			csrfToken:           csrfToken,
			wantCode:            http.StatusUnprocessableEntity,
			wantFormTag:         formTag,
		},
		{
			name:                "Empty email",
			userName:            validName,
			userEmail:           "",
			userPassword:        validPassword,
			userPasswordConfirm: validPassword,
			csrfToken:           csrfToken,
			wantCode:            http.StatusUnprocessableEntity,
			wantFormTag:         formTag,
		},
		{
			name:                "Empty password",
			userName:            validName,
			userEmail:           validEmail,
			userPassword:        "",
			userPasswordConfirm: "",
			csrfToken:           csrfToken,
			wantCode:            http.StatusUnprocessableEntity,
			wantFormTag:         formTag,
		},
		{
			name:                "Invalid email",
			userName:            validName,
			userEmail:           "bob@example.",
			userPassword:        validPassword,
			userPasswordConfirm: validPassword,
			csrfToken:           csrfToken,
			wantCode:            http.StatusUnprocessableEntity,
			wantFormTag:         formTag,
		},
		{
			name:                "Short password",
			userName:            validName,
			userEmail:           validEmail,
			userPassword:        "TooShrt",
			userPasswordConfirm: "TooShrt",
			csrfToken:           csrfToken,
			wantCode:            http.StatusUnprocessableEntity,
			wantFormTag:         formTag,
		},
		{
			name:                "Mismatched passwords",
			userName:            validName,
			userEmail:           validEmail,
			userPassword:        validPassword,
			userPasswordConfirm: "dr0W55aP",
			csrfToken:           csrfToken,
			wantCode:            http.StatusUnprocessableEntity,
			wantFormTag:         formTag,
		},
		{
			name:         "Duplicate email",
			userName:     validName,
			userEmail:    "alice@example.com",
			userPassword: validPassword,
			csrfToken:    csrfToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("password_confirm", tt.userPasswordConfirm)
			form.Add("diving_since", "1970-01-01T00:00:00Z")
			form.Add("dive_number_offset", "0")
			form.Add("default_diving_country_id", "1")
			form.Add("default_diving_tz", "Europe/London")
			form.Add("dark_mode", "true")
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/sign-up", form)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}
}
