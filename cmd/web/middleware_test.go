package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m5lapp/divesite-monolith/internal/assert"
)

func TestCommonHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	app := newTestApplication(t)
	app.commonHeaders(next).ServeHTTP(rr, r)

	rs := rr.Result()

	cspStrings := []string{
		"default-src 'self';", "script-src 'self'",
		"'sha256-ieoeWczDHkReVBsRBqaal5AFMlBtNjMzgwKvLqi/tSU='",
		"'nonce-", "style-src 'self' fonts.googleapis.com",
		"'sha256-3qgWK6nKtgtnBw4V9Vg+RnfoB1A6tJjVUiQLjuzRLhk='",
		"'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU='",
		"'sha256-9hSz5kaPZSZEIeSna4lh0Z9fYGdrbPwEsbTz0eUERAE='",
		"'sha256-m5kyiLZaikGV6b9HEBNQjwaI5HR973UB5DpBM4slRB4='",
		"'sha256-MveBq/V9Jvy+BM7/nmXo1RyNEavRcu+edYderaPfCOk='",
		"'unsafe-hashes'; font-src fonts.gstatic.com",
	}

	for _, cspString := range cspStrings {
		assert.StringContains(t, rs.Header.Get("Content-Security-Policy"), cspString)
	}

	assert.Equal(t, rs.Header.Get("Referrer-Policy"), "origin-when-cross-origin")
	assert.Equal(t, rs.Header.Get("X-Content-Type-Options"), "nosniff")
	assert.Equal(t, rs.Header.Get("X-Frame-Options"), "deny")
	assert.Equal(t, rs.Header.Get("X-XSS-Protection"), "0")
	assert.Equal(t, rs.Header.Get("Server"), "Go")
	assert.Equal(t, rs.StatusCode, http.StatusOK)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	assert.Equal(t, string(body), "OK")
}
