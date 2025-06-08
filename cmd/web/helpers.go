package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"github.com/m5lapp/divesite-monolith/internal/models"
)

func (app *app) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *app) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.log.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

func (app *app) decodePOSTForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

func (app *app) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}

func (app *app) newTemplateData(r *http.Request) (templateData, error) {
	user := models.AnonymousUser
	if app.contextGetIsAuthenticated(r) {
		user = app.contextGetUser(r)
	}

	agencies, err := app.agencies.List()
	if err != nil {
		return templateData{}, fmt.Errorf("could not fetch agency list for template: %w", err)
	}

	buddies, err := app.buddies.ListAll(user.ID)
	if err != nil {
		return templateData{}, fmt.Errorf("could not fetch buddy list for template: %w", err)
	}

	buddyRoles, err := app.buddyRoles.List()
	if err != nil {
		return templateData{}, fmt.Errorf("could not fetch buddy role list for template: %w", err)
	}

	countries, err := app.countries.List()
	if err != nil {
		return templateData{}, fmt.Errorf("could not fetch country list for template: %w", err)
	}

	agencyCourses, err := app.agencyCourses.List()
	if err != nil {
		return templateData{}, fmt.Errorf(
			"could not fetch agency course list for template: %w",
			err,
		)
	}

	currencies, err := app.currencies.List()
	if err != nil {
		return templateData{}, fmt.Errorf("could not fetch currency list for template: %w", err)
	}

	operators, err := app.operators.ListAll()
	if err != nil {
		return templateData{}, fmt.Errorf("could not fetch operator list for template: %w", err)
	}

	operatorTypes, err := app.operatorTypes.List()
	if err != nil {
		return templateData{}, fmt.Errorf(
			"could not fetch operator type list for template: %w",
			err,
		)
	}

	waterBodies, err := app.waterBodies.List()
	if err != nil {
		return templateData{}, fmt.Errorf("could not fetch water body list for template: %w", err)
	}

	waterTypes, err := app.waterTypes.List()
	if err != nil {
		return templateData{}, fmt.Errorf("could not fetch water type list for template: %w", err)
	}

	data := templateData{
		CSRFToken:       nosurf.Token(r),
		Agencies:        agencies,
		AgencyCourses:   agencyCourses,
		Buddies:         buddies,
		BuddyRoles:      buddyRoles,
		Countries:       countries,
		Currencies:      currencies,
		CurrentYear:     time.Now().Year(),
		DarkMode:        true,
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		NoValidate:      os.Getenv("DIVESITE_NOVALIDATE") == "true",
		Operators:       operators,
		OperatorTypes:   operatorTypes,
		User:            *user,
		WasPosted:       r.Method == http.MethodPost,
		WaterBodies:     waterBodies,
		WaterTypes:      waterTypes,
	}

	return data, nil
}

func (app *app) readInt(qs url.Values, key string, defaultValue int) int {
	value := qs.Get(key)

	if value == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return i
}

func (app *app) render(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	pageName string,
	data templateData,
) {
	ts, ok := app.templateCache[pageName]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", pageName)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}
