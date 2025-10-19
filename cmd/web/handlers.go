package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
	"github.com/m5lapp/divesite-monolith/internal/validator"
)

type userRegistrationForm struct {
	Name                   string          `form:"name"`
	Email                  string          `form:"email"`
	Password               string          `form:"password"`
	PasswordConfirm        string          `form:"password_confirm"`
	DivingSince            time.Time       `form:"diving_since"`
	DiveNumberOffset       int             `form:"dive_number_offset"`
	DefaultDivingCountryID int             `form:"default_diving_country_id"`
	DefaultDivingTZ        models.TimeZone `form:"default_diving_tz"`
	DarkMode               bool            `form:"dark_mode"`
	validator.Validator    `form:"-"`
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *app) userCreateGET(w http.ResponseWriter, r *http.Request) {
	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	defaultTZ, _ := models.NewTimeZone("Etc/UTC")
	data.Form = userRegistrationForm{DefaultDivingTZ: defaultTZ, DarkMode: true}

	app.render(w, r, http.StatusOK, "register.tmpl", data)
}

func (app *app) userCreatePOST(w http.ResponseWriter, r *http.Request) {
	form := &userRegistrationForm{}
	err := app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding user registration form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a valid email address",
	)

	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(
		validator.MinChars(form.Password, 8),
		"password",
		"This field must be at least 8 characters long",
	)
	form.CheckField(
		form.PasswordConfirm == form.Password,
		"password_confirm",
		"This field must match the password field",
	)

	earliestDivingSince := time.Date(1960, time.January, 1, 0, 0, 0, 0, time.UTC)
	latestDivingSince := time.Now()
	form.CheckField(
		validator.TimeBetween(form.DivingSince, earliestDivingSince, latestDivingSince),
		"diving_since",
		"This field must be between 1960-01-01 and today",
	)

	form.CheckField(
		validator.NumBetween[int](form.DiveNumberOffset, 0, 10_000),
		"dive_number_offset",
		"This field must be between 0 and 10,000 inclusive",
	)

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !form.Valid() {
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "register.tmpl", data)
		return
	}

	err = app.users.Insert(
		form.Name,
		form.Email,
		form.Password,
		form.DivingSince,
		form.DiveNumberOffset,
		form.DefaultDivingCountryID,
		form.DefaultDivingTZ,
		form.DarkMode,
	)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "This email is already registered")
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "register.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flashSuccess", "Sign up successful, please log in.")
	http.Redirect(w, r, "/user/log-in", http.StatusSeeOther)
}

type userLogInForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `       form:"-"`
}

func (app *app) userLogInGET(w http.ResponseWriter, r *http.Request) {
	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data.Form = userLogInForm{}

	app.render(w, r, http.StatusOK, "log_in.tmpl", data)
}

func (app *app) userLogInPOST(w http.ResponseWriter, r *http.Request) {
	var form userLogInForm

	err := app.decodePOSTForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a valid email address",
	)
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !form.Valid() {
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "log_in.tmpl", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "log_in.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	redirectPath := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogIn")
	if redirectPath == "" {
		redirectPath = "/"
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, redirectPath, http.StatusSeeOther)
}

func (app *app) userLogOutPOST(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flashInfo", "You have been logged out.")

	http.Redirect(w, r, "/user/log-in", http.StatusSeeOther)
}

func (app *app) home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/pages/home.tmpl",
		"./ui/html/partials/nav.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

type diveSiteForm struct {
	ID                  int             `form:"-"`
	Version             int             `form:"version"`
	Name                string          `form:"name"`
	AltName             string          `form:"alt_name"`
	Location            string          `form:"location"`
	Region              string          `form:"region"`
	CountryID           int             `form:"country"`
	TimeZone            models.TimeZone `form:"timezone"`
	Latitude            *float64        `form:"latitude"`
	Longitude           *float64        `form:"longitude"`
	WaterBodyID         int             `form:"water_body"`
	WaterTypeID         int             `form:"water_type"`
	Altitude            int             `form:"altitude"`
	MaxDepth            *float64        `form:"max_depth"`
	Notes               string          `form:"notes"`
	Rating              *int            `form:"rating"`
	validator.Validator `form:"-"`
}

func (ds *diveSiteForm) Validate() {
	ds.CheckField(validator.NotBlank(ds.Name), "name", "This field cannot be blank")
	ds.CheckField(
		validator.MaxChars(ds.Name, 256),
		"name",
		"This field cannot be more than 256 characters long",
	)

	ds.CheckField(
		validator.MaxChars(ds.AltName, 256),
		"alt_name",
		"This field cannot be more than 256 characters long",
	)

	ds.CheckField(validator.NotBlank(ds.Location), "location", "This field cannot be blank")
	ds.CheckField(
		validator.MaxChars(ds.Location, 256),
		"location",
		"This field cannot be more than 256 characters long",
	)

	ds.CheckField(
		validator.MaxChars(ds.Region, 256),
		"region",
		"This field cannot be more than 256 characters long",
	)

	if ds.Latitude != nil {
		ds.CheckField(
			*ds.Latitude >= -90 && *ds.Latitude <= 90,
			"latitude",
			"This field must be between -90 and 90 inclusive",
		)
	}
	if ds.Longitude != nil {
		ds.CheckField(
			*ds.Longitude >= -180 && *ds.Longitude <= 180,
			"longitude",
			"This field must be between -180 and 180 inclusive",
		)
	}

	ds.CheckField(
		ds.Altitude >= -422 && ds.Altitude <= 7000,
		"altitude",
		"This field must be between -422 and 7,000 inclusive",
	)

	if ds.MaxDepth != nil {
		ds.CheckField(
			*ds.MaxDepth >= 4 && *ds.MaxDepth <= 350,
			"max_depth",
			"This field must be between 4 and 350 inclusive",
		)
	}

	ds.CheckField(
		validator.MaxChars(ds.Notes, 65536),
		"notes",
		"This field cannot be more than 65,536 characters long",
	)

	if ds.Rating != nil {
		ds.CheckField(
			*ds.Rating >= 0 && *ds.Rating <= 10,
			"rating",
			"This field must be between 0 and 10 inclusive",
		)
	}
}

func (app *app) diveSiteCreateGET(w http.ResponseWriter, r *http.Request) {
	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	data.Form = diveSiteForm{
		CountryID:   user.DefaultDivingCountryID,
		TimeZone:    user.DefaultDivingTZ,
		WaterBodyID: 1,
		WaterTypeID: 1,
	}

	app.render(w, r, http.StatusOK, "dive_site/new.tmpl", data)
}

func (app *app) diveSiteCreatePOST(w http.ResponseWriter, r *http.Request) {
	form := &diveSiteForm{}
	err := app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding dive site form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Validate()
	if !form.Valid() {
		data, err := app.newTemplateData(r)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "dive_site/new.tmpl", data)
		return
	}

	id, err := app.diveSites.Insert(
		app.contextGetUser(r).ID,
		form.Name,
		form.AltName,
		form.Location,
		form.Region,
		form.CountryID,
		form.TimeZone,
		form.Latitude,
		form.Longitude,
		form.WaterBodyID,
		form.WaterTypeID,
		form.Altitude,
		form.MaxDepth,
		form.Notes,
		form.Rating,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flashSuccess", "Dive site added successfully.")

	nextUrl := fmt.Sprintf("/log-book/dive-site/view/%d", id)
	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}

func (app *app) diveSiteUpdateGET(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	userID := app.contextGetUser(r).ID

	diveSite, err := app.diveSites.GetOneByID(id, userID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Form = diveSiteForm{
		ID:          id,
		Version:     diveSite.Version,
		Name:        diveSite.Name,
		AltName:     diveSite.AltName,
		Location:    diveSite.Location,
		Region:      diveSite.Region,
		CountryID:   diveSite.Country.ID,
		TimeZone:    diveSite.TimeZone,
		Latitude:    diveSite.Latitude,
		Longitude:   diveSite.Longitude,
		WaterBodyID: diveSite.WaterBody.ID,
		WaterTypeID: diveSite.WaterType.ID,
		Altitude:    diveSite.Altitude,
		MaxDepth:    diveSite.MaxDepth,
		Notes:       diveSite.Notes,
		Rating:      diveSite.Rating,
	}

	app.render(w, r, http.StatusOK, "dive_site/new.tmpl", data)
}

func (app *app) diveSiteUpdatePOST(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	form := &diveSiteForm{}
	err = app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding dive site form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Validate()
	if !form.Valid() {
		data, err := app.newTemplateData(r)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "dive_site/new.tmpl", data)
		return
	}

	err = app.diveSites.Update(
		id,
		form.Version,
		form.Name,
		form.AltName,
		form.Location,
		form.Region,
		form.CountryID,
		form.TimeZone,
		form.Latitude,
		form.Longitude,
		form.WaterBodyID,
		form.WaterTypeID,
		form.Altitude,
		form.MaxDepth,
		form.Notes,
		form.Rating,
	)
	if err != nil {
		switch err {
		case models.ErrUpdateConflict:
			msg := `The dive site was already updated by another user, please
                    make your changes again.`
			app.sessionManager.Put(r.Context(), "flashError", msg)
			nextUrl := fmt.Sprintf("/log-book/dive-site/edit/%d", id)
			http.Redirect(w, r, nextUrl, http.StatusSeeOther)
		case models.ErrNoRecord:
			msg := `The dive site you are trying to change does not exist or
                    you do not have permission to edit it. Possibly it was
                    deleted by another user.`
			app.sessionManager.Put(r.Context(), "flashError", msg)
			http.Redirect(w, r, "/log-book/dive-site", http.StatusSeeOther)
		default:
			app.serverError(w, r, err)
		}

		return
	}

	msg := "Dive site " + form.Name + " has been updated successfully."
	app.sessionManager.Put(r.Context(), "flashSuccess", msg)

	nextUrl := fmt.Sprintf("/log-book/dive-site/view/%d", id)
	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}

func (app *app) diveSiteList(w http.ResponseWriter, r *http.Request) {
	const defaultPageSize = 20

	page := app.readInt(r.URL.Query(), "page", 1)
	pageSize := app.readInt(r.URL.Query(), "page_size", defaultPageSize)

	pager := models.NewPager(page, pageSize, defaultPageSize)
	userID := app.contextGetUser(r).ID

	diveSites, pageData, err := app.diveSites.List(userID, pager, models.SortDiveSiteDefault)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data.DiveSites = diveSites
	data.PageData = pageData

	app.render(w, r, http.StatusOK, "dive_site/list.tmpl", data)
}

func (app *app) diveSiteGET(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	userID := app.contextGetUser(r).ID

	diveSite, err := app.diveSites.GetOneByID(id, userID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	pager := models.NewPager(1, 10, 10)
	filter := models.DiveFilter{DiveSiteID: id}
	dives, _, err := app.dives.List(userID, pager, filter, models.SortDiveDefault)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data.DiveSite = diveSite
	data.Dives = dives

	app.render(w, r, http.StatusOK, "dive_site/view.tmpl", data)
}

type operatorForm struct {
	Name                string `form:"name"`
	OperatorTypeID      int    `form:"operator_type_id"`
	Street              string `form:"street"`
	Suburb              string `form:"suburb"`
	State               string `form:"state"`
	Postcode            string `form:"postcode"`
	CountryID           int    `form:"country"`
	WebsiteURL          string `form:"website_url"`
	EmailAddress        string `form:"email_address"`
	PhoneNumber         string `form:"phone_number"`
	Comments            string `form:"comments"`
	validator.Validator `       form:"-"`
}

func (app *app) operatorCreateGET(w http.ResponseWriter, r *http.Request) {
	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Form = operatorForm{}
	app.render(w, r, http.StatusOK, "operator/new.tmpl", data)
}

func (app *app) operatorCreatePOST(w http.ResponseWriter, r *http.Request) {
	form := &operatorForm{}
	err := app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding operator form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	maxCharsErrMsg := "This field cannot be more than %d characters long"

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")

	form.CheckField(form.OperatorTypeID > 0, "operator_type_id", "This field must be selected")

	form.CheckField(
		validator.MaxChars(form.Street, 256),
		"street",
		fmt.Sprintf(maxCharsErrMsg, 256),
	)

	form.CheckField(
		validator.MaxChars(form.Suburb, 256),
		"suburb",
		fmt.Sprintf(maxCharsErrMsg, 256),
	)

	form.CheckField(
		validator.MaxChars(form.Street, 256),
		"street",
		fmt.Sprintf(maxCharsErrMsg, 256),
	)

	form.CheckField(
		validator.MaxChars(form.Postcode, 16),
		"postcode",
		fmt.Sprintf(maxCharsErrMsg, 16),
	)

	form.CheckField(form.CountryID > 0, "country_id", "This field must be selected")

	form.CheckField(
		form.WebsiteURL == "" || validator.IsHTTPURL(form.WebsiteURL),
		"website_url",
		"This field must be a valid HTTP or HTTPS URL",
	)
	form.CheckField(
		validator.MaxChars(form.WebsiteURL, 2048),
		"website_url",
		fmt.Sprintf(maxCharsErrMsg, 2048),
	)

	form.CheckField(
		validator.MaxChars(form.EmailAddress, 254),
		"email_address",
		fmt.Sprintf(maxCharsErrMsg, 254),
	)
	form.CheckField(
		form.EmailAddress == "" || validator.Matches(form.EmailAddress, validator.EmailRX),
		"email_address",
		"This field must be a valid email address",
	)

	form.CheckField(
		validator.MaxChars(form.PhoneNumber, 32),
		"phone_number",
		fmt.Sprintf(maxCharsErrMsg, 32),
	)

	form.CheckField(
		validator.MaxChars(form.Comments, 4096),
		"comments",
		fmt.Sprintf(maxCharsErrMsg, 4096),
	)

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !form.Valid() {
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "operator/new.tmpl", data)
		return
	}

	_, err = app.operators.Insert(
		app.contextGetUser(r).ID,
		form.OperatorTypeID,
		form.Name,
		form.Street,
		form.Suburb,
		form.State,
		form.Postcode,
		form.CountryID,
		form.WebsiteURL,
		form.EmailAddress,
		form.PhoneNumber,
		form.Comments,
	)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "This operator already exists")
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "operator/new.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flashSuccess", "Dive operator added successfully.")
	http.Redirect(w, r, "/operator/", http.StatusSeeOther)
}

func (app *app) operatorList(w http.ResponseWriter, r *http.Request) {
	const defaultPageSize = 20

	page := app.readInt(r.URL.Query(), "page", 1)
	pageSize := app.readInt(r.URL.Query(), "page_size", defaultPageSize)

	pager := models.NewPager(page, pageSize, defaultPageSize)
	userID := app.contextGetUser(r).ID

	operators, pageData, err := app.operators.List(userID, pager)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Operators = operators
	data.PageData = pageData

	app.render(w, r, http.StatusOK, "operator/list.tmpl", data)
}

type buddyForm struct {
	Name                string `form:"name"`
	EmailAddress        string `form:"email_address"`
	PhoneNumber         string `form:"phone_number"`
	AgencyID            *int   `form:"agency_id"`
	AgencyMemberNum     string `form:"agency_member_num"`
	Notes               string `form:"notes"`
	validator.Validator `       form:"-"`
}

func (app *app) buddyList(w http.ResponseWriter, r *http.Request) {
	const defaultPageSize = 20

	page := app.readInt(r.URL.Query(), "page", 1)
	pageSize := app.readInt(r.URL.Query(), "page_size", defaultPageSize)

	pager := models.NewPager(page, pageSize, defaultPageSize)
	userID := app.contextGetUser(r).ID

	buddies, pageData, err := app.buddies.List(userID, pager)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data.Buddies = buddies
	data.PageData = pageData

	app.render(w, r, http.StatusOK, "buddy/list.tmpl", data)
}

func (app *app) buddyCreateGET(w http.ResponseWriter, r *http.Request) {
	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Form = buddyForm{}
	app.render(w, r, http.StatusOK, "buddy/new.tmpl", data)
}

func (app *app) buddyCreatePOST(w http.ResponseWriter, r *http.Request) {
	form := &buddyForm{}
	err := app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding buddy form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	maxCharsErrMsg := "This field cannot be more than %d characters long"

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(
		validator.MaxChars(form.Name, 256),
		"name",
		fmt.Sprintf(maxCharsErrMsg, 256),
	)

	form.CheckField(
		validator.MaxChars(form.EmailAddress, 254),
		"email_address",
		fmt.Sprintf(maxCharsErrMsg, 254),
	)
	form.CheckField(
		form.EmailAddress == "" || validator.Matches(form.EmailAddress, validator.EmailRX),
		"email_address",
		"This field must be a valid email address",
	)

	form.CheckField(
		validator.MaxChars(form.PhoneNumber, 32),
		"phone_number",
		fmt.Sprintf(maxCharsErrMsg, 32),
	)

	if form.AgencyID == nil {
		if form.AgencyMemberNum != "" {
			form.AddNonFieldError("Please choose an agency for the agency membership number")
		}
	} else {
		form.CheckField(*form.AgencyID > 0, "agency_id", "This field must be selected")
	}

	form.CheckField(
		validator.MaxChars(form.AgencyMemberNum, 16),
		"agency_member_num",
		fmt.Sprintf(maxCharsErrMsg, 16),
	)

	form.CheckField(
		validator.MaxChars(form.Notes, 4096),
		"notes",
		fmt.Sprintf(maxCharsErrMsg, 4096),
	)

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !form.Valid() {
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "buddy/new.tmpl", data)
		return
	}

	_, err = app.buddies.Insert(
		app.contextGetUser(r).ID,
		form.Name,
		form.EmailAddress,
		form.PhoneNumber,
		form.AgencyID,
		form.AgencyMemberNum,
		form.Notes,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flashSuccess", "Dive buddy added successfully.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type tripForm struct {
	Name                string    `form:"name"`
	StartDate           time.Time `form:"start_date"`
	EndDate             time.Time `form:"end_date"`
	Description         string    `form:"description"`
	Rating              *int      `form:"rating"`
	OperatorID          *int      `form:"operator_id"`
	PriceAmount         *float64  `form:"price"`
	CurrencyID          *int      `form:"currency_id"`
	Notes               string    `form:"notes"`
	validator.Validator `form:"-"`
}

func (app *app) tripCreateGET(w http.ResponseWriter, r *http.Request) {
	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Form = tripForm{}
	app.render(w, r, http.StatusOK, "trip/new.tmpl", data)
}

func (app *app) tripCreatePOST(w http.ResponseWriter, r *http.Request) {
	form := &tripForm{}
	err := app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding trip form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	maxCharsErrMsg := "This field cannot be more than %d characters long"

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(
		validator.MaxChars(form.Name, 256),
		"name",
		fmt.Sprintf(maxCharsErrMsg, 256),
	)

	earliestDate := time.Date(1960, time.January, 1, 0, 0, 0, 0, time.UTC)
	latestDate := time.Now().Add(365 * 24 * time.Hour)
	dateErrorMsg := "This field must be between %s and %s"
	form.CheckField(
		validator.TimeBetween(form.StartDate, earliestDate, latestDate),
		"start_date",
		fmt.Sprintf(
			dateErrorMsg,
			earliestDate.Format(time.DateOnly),
			latestDate.Format(time.DateOnly),
		),
	)
	form.CheckField(
		validator.TimeBetween(form.EndDate, earliestDate, latestDate),
		"end_date",
		fmt.Sprintf(
			dateErrorMsg,
			earliestDate.Format(time.DateOnly),
			latestDate.Format(time.DateOnly),
		),
	)
	if form.EndDate.Before(form.StartDate) {
		form.AddNonFieldError("The trip start date must be before the end date")
	}

	form.CheckField(
		validator.MaxChars(form.Description, 1024),
		"description",
		fmt.Sprintf(maxCharsErrMsg, 256),
	)

	if form.Rating != nil {
		form.CheckField(
			validator.NumBetween[int](*form.Rating, 0, 10),
			"rating",
			"This field must be between 0 and 10 inclusive",
		)
	}

	if form.OperatorID != nil {
		form.CheckField(*form.OperatorID > 0, "operator_id", "Select a valid operator")
	}

	if form.PriceAmount != nil {
		form.CheckField(
			validator.NumBetween[float64](*form.PriceAmount, 0.0, 9_999_999_999.999),
			"price",
			"This field must be between 0.0 and 9,999,999,999.99 inclusive",
		)

		if form.CurrencyID == nil {
			form.AddFieldError("currency_id", "A currency must be selected for the price")
		}
	}

	if form.CurrencyID != nil {
		form.CheckField(*form.CurrencyID > 0, "currency_id", "Select a valid currency")

		if form.PriceAmount == nil {
			form.AddFieldError(
				"price",
				"A price must be entered for the currency",
			)
		}
	}

	form.CheckField(
		validator.MaxChars(form.Notes, 4096),
		"notes",
		fmt.Sprintf(maxCharsErrMsg, 4096),
	)

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !form.Valid() {
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "trip/new.tmpl", data)
		return
	}

	_, err = app.trips.Insert(
		app.contextGetUser(r).ID,
		form.Name,
		form.StartDate,
		form.EndDate,
		form.Description,
		form.Rating,
		form.OperatorID,
		form.PriceAmount,
		form.CurrencyID,
		form.Notes,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flashSuccess", "Dive trip added successfully.")
	http.Redirect(w, r, "/trip/", http.StatusSeeOther)
}

func (app *app) tripList(w http.ResponseWriter, r *http.Request) {
	const defaultPageSize = 20

	page := app.readInt(r.URL.Query(), "page", 1)
	pageSize := app.readInt(r.URL.Query(), "page_size", defaultPageSize)

	pager := models.NewPager(page, pageSize, defaultPageSize)
	userID := app.contextGetUser(r).ID

	trips, pageData, err := app.trips.List(userID, pager)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Trips = trips
	data.PageData = pageData

	app.render(w, r, http.StatusOK, "trip/list.tmpl", data)
}

type certificationForm struct {
	CourseID            int       `form:"course_id"`
	StartDate           time.Time `form:"start_date"`
	EndDate             time.Time `form:"end_date"`
	OperatorID          int       `form:"operator_id"`
	InstructorID        int       `form:"instructor_id"`
	PriceAmount         *float64  `form:"price"`
	CurrencyID          *int      `form:"currency_id"`
	Rating              *int      `form:"rating"`
	Notes               string    `form:"notes"`
	validator.Validator `form:"-"`
}

func (app *app) certificationCreateGET(w http.ResponseWriter, r *http.Request) {
	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Form = certificationForm{}
	app.render(w, r, http.StatusOK, "certification/new.tmpl", data)
}

func (app *app) certificationCreatePOST(w http.ResponseWriter, r *http.Request) {
	form := &certificationForm{}
	err := app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding certification form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	maxCharsErrMsg := "This field cannot be more than %d characters long"

	form.CheckField(form.CourseID > 0, "course_id", "Select a valid course")

	earliestDate := time.Date(1960, time.January, 1, 0, 0, 0, 0, time.UTC)
	latestDate := time.Now().Add(365 * 24 * time.Hour)
	dateErrorMsg := "This field must be between %s and %s"
	form.CheckField(
		validator.TimeBetween(form.StartDate, earliestDate, latestDate),
		"start_date",
		fmt.Sprintf(
			dateErrorMsg,
			earliestDate.Format(time.DateOnly),
			latestDate.Format(time.DateOnly),
		),
	)
	form.CheckField(
		validator.TimeBetween(form.EndDate, earliestDate, latestDate),
		"end_date",
		fmt.Sprintf(
			dateErrorMsg,
			earliestDate.Format(time.DateOnly),
			latestDate.Format(time.DateOnly),
		),
	)
	if form.EndDate.Before(form.StartDate) {
		form.AddNonFieldError("The certification start date must be before the end date")
	}

	form.CheckField(form.OperatorID > 0, "operator_id", "Select a valid operator")

	form.CheckField(form.InstructorID > 0, "instructor_id", "Select a valid instructor")

	if form.PriceAmount != nil {
		form.CheckField(
			validator.NumBetween[float64](*form.PriceAmount, 0.0, 9_999_999_999.999),
			"price",
			"This field must be between 0.0 and 9,999,999,999.99 inclusive",
		)

		if form.CurrencyID == nil {
			form.AddFieldError("currency_id", "A currency must be selected for the price")
		}
	}

	if form.CurrencyID != nil {
		form.CheckField(*form.CurrencyID > 0, "currency_id", "Select a valid currency")

		if form.PriceAmount == nil {
			form.AddFieldError(
				"price",
				"A price must be entered for the currency",
			)
		}
	}

	if form.Rating != nil {
		form.CheckField(
			validator.NumBetween[int](*form.Rating, 0, 10),
			"rating",
			"This field must be between 0 and 10 inclusive",
		)
	}

	form.CheckField(
		validator.MaxChars(form.Notes, 4096),
		"notes",
		fmt.Sprintf(maxCharsErrMsg, 4096),
	)

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !form.Valid() {
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "certification/new.tmpl", data)
		return
	}

	_, err = app.certifications.Insert(
		app.contextGetUser(r).ID,
		form.CourseID,
		form.StartDate,
		form.EndDate,
		form.OperatorID,
		form.InstructorID,
		form.PriceAmount,
		form.CurrencyID,
		form.Rating,
		form.Notes,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flashSuccess", "Dive certification added successfully.")
	http.Redirect(w, r, "/certification/", http.StatusSeeOther)
}

func (app *app) certificationList(w http.ResponseWriter, r *http.Request) {
	const defaultPageSize = 20

	page := app.readInt(r.URL.Query(), "page", 1)
	pageSize := app.readInt(r.URL.Query(), "page_size", defaultPageSize)

	pager := models.NewPager(page, pageSize, defaultPageSize)
	userID := app.contextGetUser(r).ID

	certs, pageData, err := app.certifications.List(userID, pager)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Certifications = certs
	data.PageData = pageData

	app.render(w, r, http.StatusOK, "certification/list.tmpl", data)
}

type diveForm struct {
	ID                  int       `form:"-"`
	Version             int       `form:"version"`
	Number              int       `form:"number"`
	Activity            string    `form:"activity"`
	DiveSiteID          int       `form:"dive_site_id"`
	OperatorID          *int      `form:"operator_id"`
	PriceAmount         *float64  `form:"price_amount"`
	CurrencyID          *int      `form:"currency_id"`
	TripID              *int      `form:"trip_id"`
	CertificationID     *int      `form:"certification_id"`
	DateTimeIn          time.Time `form:"date_time_in"`
	MaxDepth            float64   `form:"max_depth"`
	AvgDepth            *float64  `form:"avg_depth"`
	BottomTimeMins      int       `form:"bottom_time"`
	SafetyStopMins      *int      `form:"safety_stop"`
	WaterTemp           *int      `form:"water_temp"`
	AirTemp             *int      `form:"air_temp"`
	Visibility          *float64  `form:"visibility"`
	CurrentID           *int      `form:"current_id"`
	WavesID             *int      `form:"waves_id"`
	BuddyID             *int      `form:"buddy_id"`
	BuddyRoleID         *int      `form:"buddy_role_id"`
	Weight              *float64  `form:"weight"`
	WeightNotes         string    `form:"weight_notes"`
	EquipmentIDs        []int     `form:"equipment_ids"`
	EquipmentNotes      string    `form:"equipment_notes"`
	TankConfigurationID int       `form:"tank_configuration_id"`
	TankMaterialID      int       `form:"tank_material_id"`
	TankVolume          float64   `form:"tank_volume"`
	GasMixID            int       `form:"gas_mix_id"`
	FO2                 float64   `form:"fo2"`
	PressureIn          *int      `form:"pressure_in"`
	PressureOut         *int      `form:"pressure_out"`
	GasMixNotes         string    `form:"gas_mix_notes"`
	EntryPointID        int       `form:"entry_point_id"`
	Rating              *int      `form:"rating"`
	PropertyIDs         []int     `form:"property_ids"`
	Notes               string    `form:"notes"`
	validator.Validator `form:"-"`
}

func diveFormFromDive(dive models.Dive) diveForm {
	form := diveForm{
		ID:                  dive.OwnerID,
		Version:             dive.Version,
		Number:              dive.Number,
		Activity:            dive.Activity,
		DiveSiteID:          dive.DiveSite.ID,
		DateTimeIn:          dive.DateTimeIn,
		MaxDepth:            dive.MaxDepth,
		AvgDepth:            dive.AvgDepth,
		BottomTimeMins:      int(dive.BottomTime.Minutes()),
		WaterTemp:           dive.WaterTemp,
		AirTemp:             dive.AirTemp,
		Visibility:          dive.Visibility,
		Weight:              dive.Weight,
		WeightNotes:         dive.WeightNotes,
		EquipmentNotes:      dive.EquipmentNotes,
		TankConfigurationID: dive.TankConfiguration.ID,
		TankMaterialID:      dive.TankMaterial.ID,
		TankVolume:          dive.TankVolume,
		GasMixID:            dive.GasMix.ID,
		FO2:                 dive.FO2,
		PressureIn:          dive.PressureIn,
		PressureOut:         dive.PressureOut,
		GasMixNotes:         dive.GasMixNotes,
		EntryPointID:        dive.EntryPoint.ID,
		Rating:              dive.Rating,
		Notes:               dive.Notes,
	}

	if dive.Operator != nil {
		form.OperatorID = &dive.Operator.ID
	}

	if dive.Price != nil {
		form.PriceAmount = &dive.Price.Amount
		form.CurrencyID = &dive.Price.Currency.ID
	}

	if dive.Trip != nil {
		form.TripID = &dive.Trip.ID
	}

	if dive.Certification != nil {
		form.CertificationID = &dive.Certification.ID
	}

	if dive.SafetyStop != nil {
		ss := int(dive.SafetyStop.Minutes())
		form.SafetyStopMins = &ss
	}

	if dive.Current != nil {
		form.CurrentID = &dive.Current.ID
	}

	if dive.Waves != nil {
		form.WavesID = &dive.Waves.ID
	}

	if dive.Buddy != nil {
		form.BuddyID = &dive.Buddy.ID
	}

	if dive.BuddyRole != nil {
		form.BuddyRoleID = &dive.BuddyRole.ID
	}

	var equipmentIDs []int
	for _, item := range dive.Equipment {
		equipmentIDs = append(equipmentIDs, item.ID)
	}
	form.EquipmentIDs = equipmentIDs

	var propertyIDs []int
	for _, property := range dive.Properties {
		propertyIDs = append(propertyIDs, property.ID)
	}
	form.PropertyIDs = propertyIDs

	return form
}

func (app *app) addStaticdataToDiveForm(r *http.Request, data *templateData) error {
	user := models.AnonymousUser
	if app.contextGetIsAuthenticated(r) {
		user = app.contextGetUser(r)
	}

	certifications, err := app.certifications.ListAll(user.ID)
	if err != nil {
		return fmt.Errorf("could not fetch certifications list: %w", err)
	}
	data.Certifications = certifications

	currents, err := app.currents.List(false)
	if err != nil {
		return fmt.Errorf("could not fetch currents list: %w", err)
	}
	data.Currents = currents

	diveSites, err := app.diveSites.ListAll(user.ID)
	if err != nil {
		return fmt.Errorf("could not fetch dive sites list: %w", err)
	}
	data.DiveSites = diveSites

	diveProperties, err := app.diveProperties.List()
	if err != nil {
		return fmt.Errorf("could not fetch dive properties list: %w", err)
	}
	data.DiveProperties = diveProperties

	entryPoints, err := app.entryPoints.List(false)
	if err != nil {
		return fmt.Errorf("could not fetch entry points list: %w", err)
	}
	data.EntryPoints = entryPoints

	equipment, err := app.equipment.List()
	if err != nil {
		return fmt.Errorf("could not fetch equipment list: %w", err)
	}
	data.Equipment = equipment

	gasMixes, err := app.gasMixes.List(true)
	if err != nil {
		return fmt.Errorf("could not fetch gas mixes list: %w", err)
	}
	data.GasMixes = gasMixes

	tankConfigurations, err := app.tankConfigurations.List(false)
	if err != nil {
		return fmt.Errorf("could not fetch tank configurations list: %w", err)
	}
	data.TankConfigurations = tankConfigurations

	tankMaterials, err := app.tankMaterials.List(false)
	if err != nil {
		return fmt.Errorf("could not fetch tank materials list: %w", err)
	}
	data.TankMaterials = tankMaterials

	trips, err := app.trips.ListAll(user.ID)
	if err != nil {
		return fmt.Errorf("could not fetch trips list: %w", err)
	}
	data.Trips = trips

	waves, err := app.waves.List(false)
	if err != nil {
		return fmt.Errorf("could not fetch waves list: %w", err)
	}
	data.Waves = waves

	return nil
}

func (app *app) validateDiveForm(f *diveForm) error {
	f.CheckField(
		f.Number >= 1 && f.Number <= 100_000,
		"number",
		"This field must be between 1 and 100,000 inclusive",
	)

	f.CheckField(validator.NotBlank(f.Activity), "activity", "This field cannot be blank")
	f.CheckField(
		validator.MaxChars(f.Activity, 256),
		"activity",
		"This field cannot be more than 256 characters long",
	)

	exists, err := app.diveSites.Exists(f.DiveSiteID)
	if err != nil {
		return err
	}
	f.CheckField(exists, "dive_site_id", "You must select a valid dive site")

	if f.OperatorID != nil {
		exists, err := app.operators.Exists(*f.OperatorID)
		if err != nil {
			return err
		}
		f.CheckField(exists, "operator_id", "Invalid dive operator selected")
	}

	if f.PriceAmount != nil {
		f.CheckField(
			validator.NumBetween(*f.PriceAmount, 0.0, 9_999_999_999.999),
			"price_amount",
			"This field must be between 0.0 and 9,999,999,999.99 inclusive",
		)

		if f.CurrencyID == nil {
			f.AddFieldError("currency_id", "A currency must be selected for the price")
		}
	}

	if f.CurrencyID != nil {
		f.CheckField(*f.CurrencyID > 0, "currency_id", "Select a valid currency")

		if f.PriceAmount == nil {
			f.AddFieldError("price", "A price must be entered for the currency")
		}
	}

	if f.TripID != nil {
		exists, err := app.trips.Exists(*f.TripID)
		if err != nil {
			return err
		}
		f.CheckField(exists, "trip_id", "Invalid dive trip selected")
	}

	if f.CertificationID != nil {
		exists, err := app.certifications.Exists(*f.CertificationID)
		if err != nil {
			return err
		}
		f.CheckField(exists, "certification_id", "Invalid certification course selected")
	}

	earliestDiveDate := time.Date(1960, time.January, 1, 0, 0, 0, 0, time.UTC)
	latestDiveDate := time.Now().Add(24 * time.Hour)
	f.CheckField(
		validator.TimeBetween(f.DateTimeIn, earliestDiveDate, latestDiveDate),
		"date_time_in",
		"This field must be between 1960-01-01 and today",
	)

	f.CheckField(
		f.MaxDepth >= 4.0 && f.MaxDepth <= 350.0,
		"max_depth",
		"This field must be between 4 and 350 inclusive",
	)

	if f.AvgDepth != nil {
		f.CheckField(
			validator.NumBetween(*f.AvgDepth, 4.0, 350.0),
			"avg_depth",
			"This field must be between 4 and 350 inclusive",
		)

		f.CheckField(
			f.MaxDepth > *f.AvgDepth,
			"avg_depth",
			"Average depth must be less than the max depth",
		)
	}

	f.CheckField(
		f.BottomTimeMins >= 10 && f.BottomTimeMins <= 1440,
		"bottom_time",
		"This field must be between 10 and 1,440 minutes inclusive",
	)

	if f.SafetyStopMins != nil {
		f.CheckField(
			*f.SafetyStopMins >= 0 && *f.SafetyStopMins <= 6,
			"safety_stop",
			"This field must be between 0 and 6 minutes inclusive",
		)
	}

	if f.WaterTemp != nil {
		f.CheckField(
			*f.WaterTemp >= -3 && *f.WaterTemp <= 50,
			"water_temp",
			"This field must be between -3°C and 50°C inclusive",
		)
	}

	if f.AirTemp != nil {
		f.CheckField(
			*f.AirTemp >= -90 && *f.AirTemp <= 60,
			"air_temp",
			"This field must be between -90°C and 60°C inclusive",
		)
	}

	if f.Visibility != nil {
		f.CheckField(
			*f.Visibility >= 0.0 && *f.Visibility <= 80.0,
			"visibility",
			"This field must be between 0 and 80 metres inclusive",
		)
	}

	if f.CurrentID != nil {
		exists, err := app.currents.Exists(*f.CurrentID)
		if err != nil {
			return err
		}
		f.CheckField(exists, "current_id", "Invalid current type selected")
	}

	if f.WavesID != nil {
		exists, err := app.waves.Exists(*f.WavesID)
		if err != nil {
			return err
		}
		f.CheckField(exists, "waves_id", "Invalid wave type selected")
	}

	if f.BuddyID != nil {
		exists, err := app.buddies.Exists(*f.BuddyID)
		if err != nil {
			return err
		}
		f.CheckField(exists, "buddy_id", "Invalid dive buddy selected")
	}

	if f.BuddyRoleID != nil {
		exists, err := app.buddyRoles.Exists(*f.BuddyRoleID)
		if err != nil {
			return err
		}
		f.CheckField(exists, "buddy_role_id", "Invalid buddy role selected")

		f.CheckField(
			f.BuddyID != nil,
			"buddy_role_id",
			"A buddy must be selected when this field is selected",
		)
	}

	if f.Weight != nil {
		f.CheckField(
			*f.Weight >= 0.0 && *f.Weight <= 99.99,
			"weight",
			"This field must be between 0 and 99.99kg inclusive",
		)
	}

	f.CheckField(
		validator.MaxChars(f.WeightNotes, 1024),
		"weight_notes",
		"This field cannot be more than 1,024 characters long",
	)

	allExist, err := app.equipment.AllExist(f.EquipmentIDs)
	if err != nil {
		return err
	}
	f.CheckField(allExist, "equipment_ids", "Invalid equipment item(s) selected")

	f.CheckField(
		validator.MaxChars(f.EquipmentNotes, 1024),
		"equipment_notes",
		"This field cannot be more than 1,024 characters long",
	)

	exists, err = app.tankConfigurations.Exists(f.TankConfigurationID)
	if err != nil {
		return err
	}
	f.CheckField(exists, "tank_configuration_id", "Invalid tank configuration selected")

	exists, err = app.tankMaterials.Exists(f.TankMaterialID)
	if err != nil {
		return err
	}
	f.CheckField(exists, "tank_material_id", "Invalid tank material selected")

	f.CheckField(
		f.TankVolume >= 2.0 && f.TankVolume <= 22.0,
		"tank_volume",
		"This field must be between 2 and 22 litres inclusive",
	)

	gasMix, err := app.gasMixes.GetOneByID(f.GasMixID)
	if errors.Is(err, models.ErrNoRecord) {
		f.CheckField(exists, "gas_mix_id", "Invalid gas mix selected")
	} else if err != nil {
		return err
	}

	switch gasMix.Name {
	case "Air":
		f.CheckField(f.FO2 == 0.21, "fo2", "FO₂ must be 0.21 when Air is selected")
	case "Heliox":
		f.CheckField(
			validator.NumBetween(f.FO2, 0.04, 0.6),
			"fo2",
			"FO₂ must be between 0.04 and 0.6 when Heliox is selected",
		)
	case "Nitrox":
		f.CheckField(
			validator.NumBetween(f.FO2, 0.21, 0.6),
			"fo2",
			"FO₂ must be between 0.21 and 0.6 when Nitrox is selected",
		)
	case "Oxygen":
		f.CheckField(f.FO2 == 1.0, "fo2", "FO₂ must be 1.0 when Oxygen is selected")
	case "Trimix":
		f.CheckField(
			validator.NumBetween(f.FO2, 0.04, 0.6),
			"fo2",
			"FO₂ must be between 0.04 and 0.6 when Trimix is selected",
		)
	}

	// Check the FO2 values range after the individual values as this will then
	// take precedence over any gas-specific errors.
	f.CheckField(
		f.FO2 >= 0.04 && f.FO2 <= 1.0,
		"fo2",
		"This field must be between 0.04 (4%) and 1.0 (100%) inclusive",
	)

	if f.PressureIn != nil {
		f.CheckField(
			*f.PressureIn >= 150 && *f.PressureIn <= 1_000,
			"pressure_in",
			"This field must be between 150 and 1,000 bar inclusive",
		)
	}

	if f.PressureOut != nil {
		f.CheckField(
			*f.PressureOut >= 0 && *f.PressureOut <= 1_000,
			"pressure_out",
			"This field must be between 0 and 1,000 bar inclusive",
		)

		if f.PressureIn != nil {
			f.CheckField(
				*f.PressureIn > *f.PressureOut,
				"pressure_out",
				"The end pressure must be less than the starting pressure",
			)
		}
	}

	f.CheckField(
		validator.MaxChars(f.GasMixNotes, 1024),
		"gas_mix_notes",
		"This field cannot be more than 1,024 characters long",
	)

	exists, err = app.entryPoints.Exists(f.EntryPointID)
	if err != nil {
		return err
	}
	f.CheckField(exists, "entry_point_id", "Invalid entry point selected")

	if f.Rating != nil {
		f.CheckField(
			*f.Rating >= 0 && *f.Rating <= 10,
			"rating",
			"This field must be between 0 and 10 inclusive",
		)
	}

	allExist, err = app.diveProperties.AllExist(f.PropertyIDs)
	if err != nil {
		return err
	}
	f.CheckField(allExist, "property_ids", "Invalid dive properties selected")

	f.CheckField(
		validator.MaxChars(f.Notes, 65536),
		"notes",
		"This field cannot be more than 65,536 characters long",
	)

	return nil
}

func (app *app) diveCreateGET(w http.ResponseWriter, r *http.Request) {
	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Form = diveForm{
		Number:         app.contextGetUser(r).TotalDives + 1,
		MaxDepth:       5,
		BottomTimeMins: 10,
		FO2:            0.21,
		TankVolume:     11.0,
	}

	err = app.addStaticdataToDiveForm(r, &data)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed to load dive form static data: %w", err))
	}

	app.render(w, r, http.StatusOK, "dive/form.tmpl", data)
}

func (app *app) diveCreatePOST(w http.ResponseWriter, r *http.Request) {
	form := &diveForm{}
	err := app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding dive form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.validateDiveForm(form)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed to validate dive form: %w", err))
		return
	}

	if !form.Valid() {
		err = app.addStaticdataToDiveForm(r, &data)
		if err != nil {
			errMsg := "failed to load dive form static data: %w"
			app.serverError(w, r, fmt.Errorf(errMsg, err))
			return
		}

		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "dive/form.tmpl", data)
		return
	}

	var safetyStop *time.Duration
	if form.SafetyStopMins != nil {
		ss := time.Duration(*form.SafetyStopMins) * time.Minute
		safetyStop = &ss
	}

	id, err := app.dives.Insert(
		app.contextGetUser(r).ID,
		form.Number,
		form.Activity,
		form.DiveSiteID,
		form.OperatorID,
		form.PriceAmount,
		form.CurrencyID,
		form.TripID,
		form.CertificationID,
		form.DateTimeIn,
		form.MaxDepth,
		form.AvgDepth,
		time.Duration(form.BottomTimeMins)*time.Minute,
		safetyStop,
		form.WaterTemp,
		form.AirTemp,
		form.Visibility,
		form.CurrentID,
		form.WavesID,
		form.BuddyID,
		form.BuddyRoleID,
		form.Weight,
		form.WeightNotes,
		form.EquipmentIDs,
		form.EquipmentNotes,
		form.TankConfigurationID,
		form.TankMaterialID,
		form.TankVolume,
		form.GasMixID,
		form.FO2,
		form.PressureIn,
		form.PressureOut,
		form.GasMixNotes,
		form.EntryPointID,
		form.PropertyIDs,
		form.Rating,
		form.Notes,
	)
	if err != nil {
		switch err {
		case models.ErrDuplicateDiveNumber:
			form.AddFieldError("number", "A dive has already been logged with this number")
			err = app.addStaticdataToDiveForm(r, &data)
			if err != nil {
				errMsg := "failed to load dive form static data: %w"
				app.serverError(w, r, fmt.Errorf(errMsg, err))
				return
			}
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "dive/form.tmpl", data)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	nextUrl := fmt.Sprintf("/log-book/dive/view/%d", id)
	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}

func (app *app) diveUpdateGET(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	userID := app.contextGetUser(r).ID

	dive, err := app.dives.GetOneByID(userID, id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Form = diveFormFromDive(dive)

	err = app.addStaticdataToDiveForm(r, &data)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed to load dive form static data: %w", err))
	}

	app.render(w, r, http.StatusOK, "dive/form.tmpl", data)
}

func (app *app) diveUpdatePOST(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	form := &diveForm{}
	err = app.decodePOSTForm(r, form)
	if err != nil {
		app.log.Error("Error whilst decoding dive form input", "error", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.validateDiveForm(form)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed to validate dive form: %w", err))
		return
	}

	if !form.Valid() {
		err = app.addStaticdataToDiveForm(r, &data)
		if err != nil {
			errMsg := "failed to load dive form static data: %w"
			app.serverError(w, r, fmt.Errorf(errMsg, err))
			return
		}

		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "dive/form.tmpl", data)
		return
	}

	var safetyStop *time.Duration
	if form.SafetyStopMins != nil {
		ss := time.Duration(*form.SafetyStopMins) * time.Minute
		safetyStop = &ss
	}

	err = app.dives.Update(
		id,
		app.contextGetUser(r).ID,
		form.Number,
		form.Activity,
		form.DiveSiteID,
		form.OperatorID,
		form.PriceAmount,
		form.CurrencyID,
		form.TripID,
		form.CertificationID,
		form.DateTimeIn,
		form.MaxDepth,
		form.AvgDepth,
		time.Duration(form.BottomTimeMins)*time.Minute,
		safetyStop,
		form.WaterTemp,
		form.AirTemp,
		form.Visibility,
		form.CurrentID,
		form.WavesID,
		form.BuddyID,
		form.BuddyRoleID,
		form.Weight,
		form.WeightNotes,
		form.EquipmentIDs,
		form.EquipmentNotes,
		form.TankConfigurationID,
		form.TankMaterialID,
		form.TankVolume,
		form.GasMixID,
		form.FO2,
		form.PressureIn,
		form.PressureOut,
		form.GasMixNotes,
		form.EntryPointID,
		form.PropertyIDs,
		form.Rating,
		form.Notes,
	)
	if err != nil {
		switch err {
		case models.ErrDuplicateDiveNumber:
			form.AddFieldError("number", "A dive has already been logged with this number")
			err = app.addStaticdataToDiveForm(r, &data)
			if err != nil {
				errMsg := "failed to load dive form static data: %w"
				app.serverError(w, r, fmt.Errorf(errMsg, err))
				return
			}
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "dive/form.tmpl", data)
		case models.ErrNoRecord:
			msg := `The dive you are trying to change does not exist or you do
                    not have permission to edit it.`
			app.sessionManager.Put(r.Context(), "flashError", msg)
			http.Redirect(w, r, "/log-book/dive", http.StatusSeeOther)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	msg := fmt.Sprintf("Dive number %d has been updated successfully.", form.Number)
	app.sessionManager.Put(r.Context(), "flashSuccess", msg)

	nextUrl := fmt.Sprintf("/log-book/dive/view/%d", id)
	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}

func (app *app) diveGET(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	dive, err := app.dives.GetOneByID(user.ID, id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data.Dive = dive

	app.render(w, r, http.StatusOK, "dive/view.tmpl", data)
}

func (app *app) diveList(w http.ResponseWriter, r *http.Request) {
	const defaultPageSize = 20

	user := app.contextGetUser(r)
	page := app.readInt(r.URL.Query(), "page", 1)
	pageSize := app.readInt(r.URL.Query(), "page_size", defaultPageSize)

	pager := models.NewPager(page, pageSize, defaultPageSize)
	filter := models.DiveFilter{}

	records, pageData, err := app.dives.List(user.ID, pager, filter, models.SortDiveDefault)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data, err := app.newTemplateData(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data.Dives = records
	data.PageData = pageData

	app.render(w, r, http.StatusOK, "dive/list.tmpl", data)
}
