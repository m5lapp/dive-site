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

	app.sessionManager.Put(r.Context(), "flash", "Sign up successful, please log in.")
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

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *app) userLogOutPOST(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You have been logged out.")

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

	app.sessionManager.Put(r.Context(), "flash", "Dive site added successfully.")

	nextUrl := fmt.Sprintf("/log-book/dive-site/view/%d", id)
	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}

func (app *app) diveSiteList(w http.ResponseWriter, r *http.Request) {
	const defaultPageSize = 20

	page := app.readInt(r.URL.Query(), "page", 1)
	pageSize := app.readInt(r.URL.Query(), "page_size", defaultPageSize)

	filters := models.NewListFilters(page, pageSize, defaultPageSize)

	diveSites, pageData, err := app.diveSites.List(filters)
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

func (app *app) diveSiteGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	diveSite, err := app.diveSites.GetOneByID(id)
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
	data.DiveSite = diveSite

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

	app.sessionManager.Put(r.Context(), "flash", "Dive operator added successfully.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	if form.AgencyID != nil {
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

	app.sessionManager.Put(r.Context(), "flash", "Dive buddy added successfully.")
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

	app.sessionManager.Put(r.Context(), "flash", "Dive trip added successfully.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
