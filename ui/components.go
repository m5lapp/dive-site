package ui

import (
	"fmt"
	"html/template"
	"strings"

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
