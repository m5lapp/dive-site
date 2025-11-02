package ui

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/m5lapp/divesite-monolith/internal/models"
	g "maragu.dev/gomponents"
	c "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func renderGomponent(component g.Node) (template.HTML, error) {
	var htmlBuilder strings.Builder

	err := component.Render(&htmlBuilder)
	if err != nil {
		return template.HTML(""), err
	}

	// Convert the raw HTML string to a template.HTML string so that any
	// templates render it as HTML and not as an escaped string.
	return template.HTML(htmlBuilder.String()), err
}

func pageField(text, link string, linkPage, pageSize int, active, disabled bool) g.Node {
	urlPath := fmt.Sprintf("%s?page=%d&page_size=%d", link, linkPage, pageSize)

	return Li(
		c.Classes{"page-item": true, "active": active, "disabled": disabled},
		A(Class("page-link"), Href(urlPath), g.Text(text)),
	)
}

func PageControls(path string, pd models.PageData) (template.HTML, error) {
	isCurr := func(page int) bool { return pd.CurrentPage == page }

	navList := Ul(
		Class("pagination justify-content-center"),
		pageField("First", path, 1, pd.PageSize, false, isCurr(1)),
		pageField("Previous", path, pd.CurrentPage-1, pd.PageSize, false, isCurr(1)),
		pageField(
			fmt.Sprintf("Page %d of %d", pd.CurrentPage, pd.LastPage),
			path,
			pd.CurrentPage,
			pd.PageSize,
			true,
			false,
		),
		pageField("Next", path, pd.CurrentPage+1, pd.PageSize, false, isCurr(pd.LastPage)),
		pageField("Last", path, pd.LastPage, pd.PageSize, false, isCurr(pd.LastPage)),
	)

	component := Nav(
		Aria("label", "Pagination navigation"),
		Div(Class("row my-3"), navList),
	)

	return renderGomponent(component)
}

func id(name string) string {
	return "id_" + name
}

// displayName returns a field name for displaying to an end-user, defaulting to
// the field's internal name with the first character of each word upper-cased
// and underscores replaced with spaces if displayName is empty. If required is
// true, then an asterisk will be added to the end of the returned string.
func displayName(displayName, name string, required bool) string {
	if displayName == "" {
		// Upper-case the first character of each component of the field name.
		nameParts := strings.Split(name, "_")

		for i, part := range nameParts {
			r, size := utf8.DecodeRuneInString(part)
			if r != utf8.RuneError {
				nameParts[i] = string(unicode.ToUpper(r)) + part[size:]
			}
		}

		displayName = strings.Join(nameParts, " ")
	}

	if required {
		return displayName + " *"
	}

	return displayName
}

// BSBoolField returns a template.HTML string that renders a form element using
// the HTML5 checkbox input type. The switchStyle parameter determines whether
// or not the element is displayed as a Bootstrap switch-style element or as a
// traditional checkbox.
func BSBoolField(
	name, dispName, value string,
	checked, required, switchStyle bool,
	fieldErrs map[string]string,
) (template.HTML, error) {
	var fieldErr string
	fieldErr, _ = fieldErrs[name]

	component := Div(
		c.Classes{"col-sm": true},
		Div(
			c.Classes{"form-check": true, "form-switch": switchStyle},
			Label(
				Class("form-check-label"),
				For(id(name)),
				g.Text(displayName(dispName, name, required)),
			),
			Input(
				Type("checkbox"),
				ID(id(name)),
				Name(name),
				Value(value),
				Aria("described-by", id(name)+"_feedback"),
				c.Classes{
					"form-control":     true,
					"form-check-input": true,
					"is-invalid":       fieldErr != "",
				},
				g.If(checked, Checked()),
				g.If(required, Required()),
				g.If(switchStyle, Role("switch")),
			),
			g.If(
				fieldErr != "",
				Div(Class("invalid-feedback"), ID(id(name)+"_feedback"), g.Text(fieldErr)),
			),
		),
	)

	return renderGomponent(component)
}

func strToTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	if value == "now" {
		return time.Now(), nil
	}

	if strings.HasPrefix(value, "now") {
		duration, err := time.ParseDuration(value[3:])
		if err != nil {
			return time.Time{}, err
		}
		return time.Now().Add(duration), nil
	}

	if len(value) == 10 {
		return time.Parse(time.DateOnly, value)
	}

	return time.Parse(time.DateTime, value)
}

// BSDateField takes a time.Time value and returns a template.HTML string that
// renders that value in a form using the appropriate HTML5 input type; 'date'
// if the withTime field is true, or datetime-local otherwise. The minVal and
// maxVal and defaultVal parameters should be in the time.DateTime or
// time.DateOnly format, or be be "now" or "" for the zero time value.
func BSDateField(
	name, dispName, minVal, maxVal, defaultVal string,
	value time.Time,
	required, withTime bool,
	fieldErrs map[string]string,
) (template.HTML, error) {
	fieldErr, _ := fieldErrs[name]

	inputFormat := time.DateOnly
	inputType := "date"
	step := 1 // 1 day.

	if withTime {
		inputFormat = "2006-01-02T15:04"
		inputType = "datetime-local"
		step = 60 // 60 seconds.
	}

	const parseErrorMsg = "failed to parse field %s (%s) for field %s: %w"
	minTime, err := strToTime(minVal)
	if err != nil {
		return "", fmt.Errorf(parseErrorMsg, "minVal", minVal, name, err)
	}
	maxTime, err := strToTime(maxVal)
	if err != nil {
		return "", fmt.Errorf(parseErrorMsg, "maxVal", maxVal, name, err)
	}
	defaultTime, err := strToTime(defaultVal)
	if err != nil {
		return "", fmt.Errorf(parseErrorMsg, "defaultVal", defaultVal, name, err)
	}

	component := Div(
		Class("col-sm"),
		Label(Class("form-label"), For(id(name)), g.Text(displayName(dispName, name, required))),
		Input(
			Type(inputType),
			ID(id(name)),
			Name(name),
			Aria("described-by", id(name)+"_feedback"),
			c.Classes{"form-control": true, "is-invalid": fieldErr != ""},
			g.If(required, Required()),
			Min(minTime.Format(inputFormat)),
			Max(maxTime.Format(inputFormat)),
			Step(fmt.Sprintf("%d", step)),
			g.If(
				value.IsZero() == true,
				// By default, truncate the seconds to zero. This will make the
				// date picker show just hours and minutes in the browser.
				Value(defaultTime.Truncate(time.Minute).Format(inputFormat)),
			),
			g.If(value.IsZero() == false, Value(value.Format(inputFormat))),
		),
		g.If(
			fieldErr != "",
			Div(Class("invalid-feedback"), ID(id(name)+"_feedback"), g.Text(fieldErr)),
		),
	)

	return renderGomponent(component)
}

func bsNumField(
	name, dispName, min, max, step, value string,
	required bool,
	fieldErrs map[string]string,
) (template.HTML, error) {
	var fieldErr string
	fieldErr, _ = fieldErrs[name]

	component := Div(
		Class("col-sm"),
		Label(Class("form-label"), For(id(name)), g.Text(displayName(dispName, name, required))),
		Input(
			Type("number"),
			Value(value),
			ID(id(name)),
			Name(name),
			Aria("described-by", id(name)+"_feedback"),
			c.Classes{"form-control": true, "is-invalid": fieldErr != ""},
			Min(min),
			Max(max),
			Step(step),
			g.If(required, Required()),
		),
		g.If(
			fieldErr != "",
			Div(Class("invalid-feedback"), ID(id(name)+"_feedback"), g.Text(fieldErr)),
		),
	)

	return renderGomponent(component)
}

// BSNumField takes any numeric type and returns a template.HTML string that
// renders that value in a form using the HTML5 number input type.
func BSNumField[N float32 | float64 | int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64](
	name, dispName, min, max, step string,
	value N,
	required bool,
	fieldErrs map[string]string,
) (template.HTML, error) {
	var strValue string

	switch any(value).(type) {
	case float32, float64:
		strValue = strconv.FormatFloat(float64(value), 'f', -1, 64)
	default:
		strValue = fmt.Sprintf("%v", value)
	}

	return bsNumField(name, dispName, min, max, step, strValue, required, fieldErrs)
}

// BSNumFieldPtr takes any numeric pointer type and returns a template.HTML
// string that renders that value in a form using the HTML5 number input type.
func BSNumFieldPtr[N float32 | float64 | int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64](
	name, dispName, min, max, step string,
	value *N,
	required bool,
	fieldErrs map[string]string,
) (template.HTML, error) {
	var strValue string

	if value != nil {
		switch any(*value).(type) {
		case float32, float64:
			strValue = strconv.FormatFloat(float64(*value), 'f', -1, 64)
		default:
			strValue = fmt.Sprintf("%v", *value)
		}
	}

	return bsNumField(name, dispName, min, max, step, strValue, required, fieldErrs)
}

// BSTextField takes a string type and returns a template.HTML string that
// renders that value in a form using the HTML5 text input type.
func BSTextField(
	fieldType, name, dispName, value, minLength, maxLength string,
	required bool,
	fieldErrs map[string]string,
) (template.HTML, error) {
	var fieldErr string
	fieldErr, _ = fieldErrs[name]

	component := Div(
		Class("col-sm"),
		Label(Class("form-label"), For(id(name)), g.Text(displayName(dispName, name, required))),
		Input(
			Type(fieldType),
			Value(value),
			ID(id(name)),
			Name(name),
			Aria("described-by", id(name)+"_feedback"),
			c.Classes{"form-control": true, "is-invalid": fieldErr != ""},
			MinLength(minLength),
			MaxLength(maxLength),
			g.If(required, Required()),
		),
		g.If(
			fieldErr != "",
			Div(Class("invalid-feedback"), ID(id(name)+"_feedback"), g.Text(fieldErr)),
		),
	)

	return renderGomponent(component)
}
