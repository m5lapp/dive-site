package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/m5lapp/divesite-monolith/internal/models/mocks"
)

var csrfTokenRX = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="(.+)">`)

func extractCSRFToken(t *testing.T, body string) string {
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatalf("no csrf token found in body: %s", body)
	}

	return html.UnescapeString(matches[1])
}

func newTestApplication(t *testing.T) *app {
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &app{
		log:                slog.New(slog.NewTextHandler(io.Discard, nil)),
		formDecoder:        formDecoder,
		sessionManager:     sessionManager,
		templateCache:      templateCache,
		agencies:           &mocks.AgencyModel{},
		agencyCourses:      &mocks.AgencyCourseModel{},
		buddies:            &mocks.BuddyModel{},
		buddyRoles:         &mocks.BuddyRoleModel{},
		certifications:     &mocks.CertificationModel{},
		countries:          &mocks.CountryModel{},
		currencies:         &mocks.CurrencyModel{},
		currents:           &mocks.CurrentModel{},
		diveProperties:     &mocks.DivePropertyModel{},
		dives:              &mocks.DiveModel{},
		diveSites:          &mocks.DiveSiteModel{},
		entryPoints:        &mocks.EntryPointModel{},
		equipment:          &mocks.EquipmentModel{},
		gasMixes:           &mocks.GasMixModel{},
		operators:          &mocks.OperatorModel{},
		operatorTypes:      &mocks.OperatorTypeModel{},
		tankConfigurations: &mocks.TankConfigurationModel{},
		tankMaterials:      &mocks.TankMaterialModel{},
		trips:              &mocks.TripModel{},
		users:              &mocks.UserModel{},
		waterBodies:        &mocks.WaterBodyModel{},
		waterTypes:         &mocks.WaterTypeModel{},
		waves:              &mocks.WavesModel{},
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// authenticate attempts to log the user in with the given email address and
// password. If the email or password are blank, then a known default valid one
// will be used instead. If the log in process succeeds, then the log in cookie
// will get added to the testServer's cookie jar and the CSRF token will be
// returned for use in future non-safe requests.
func (ts *testServer) logIn(t *testing.T, email, password string) string {
	if email == "" {
		email = "alice@example.com"
	}
	if password == "" {
		password = "Pa55W0rd"
	}

	_, _, body := ts.get(t, "/user/log-in")
	csrfToken := extractCSRFToken(t, body)

	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	form.Add("csrf_token", csrfToken)

	code, _, _ := ts.postForm(t, "/user/log-in", form)
	if code != http.StatusSeeOther {
		err := fmt.Errorf(
			"unexpected HTTP response code (%d) when logging in as %s/%s",
			code,
			email,
			password,
		)
		t.Fatal(err)
	}

	return csrfToken
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) postForm(
	t *testing.T,
	urlPath string,
	form url.Values,
) (int, http.Header, string) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}
