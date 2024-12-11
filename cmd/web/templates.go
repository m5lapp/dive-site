package main

import (
	"errors"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/m5lapp/divesite-monolith/internal/models"
	"github.com/m5lapp/divesite-monolith/ui"
)

// https://stackoverflow.com/a/48887775/641460
func isoCountryToEmoji(code string) (string, error) {
	if len(code) != 2 {
		return "", errors.New("iso country code must be two letters")
	}

	if code[0] < 'A' || code[0] > 'Z' || code[1] < 'A' || code[1] > 'Z' {
		return "", errors.New("invalid iso country code")
	}

	rune1 := string(0x1F1E6 + rune(code[0]) - 'A')
	rune2 := string(0x1F1E6 + rune(code[1]) - 'A')

	return string(rune1 + rune2), nil
}

// getOSTimeZones searches the zoneDirs directories for files whose names are
// valid timezones suitable for passing to time.LoadLocation() and returns the
// results as a slice of strings. See:
// https://stackoverflow.com/questions/40120056/get-a-list-of-valid-time-zones-in-go
func getOSTimeZones() []string {
	var zones []string
	var zonesDirs = [3]string{
		"/usr/lib/locale/TZ/",
		"/usr/share/lib/zoneinfo/",
		"/usr/share/zoneinfo/",
	}

	for _, zoneDir := range zonesDirs {
		zones = walkTZDir(zoneDir, zones)

		for i, _ := range zones {
			zones[i] = strings.TrimPrefix(zones[i], zoneDir)

			// Check that each timezone value can be loaded successfully.
			// _, err := time.LoadLocation(zones[i])
			// if err != nil {
			// 	   Remove the timezone from the zones slice,
			// }
		}
	}

	return zones
}

func walkTZDir(path string, zones []string) []string {
	files, err := os.ReadDir(path)
	if err != nil {
		return zones
	}

	for _, file := range files {
		// The first character of every valid timezone component should be upper
		// case letter. There may be some auxiliary files that are not and
		// should therefore be excluded.
		firstRune := rune(file.Name()[0])
		if !unicode.IsUpper(firstRune) || !unicode.IsLetter(firstRune) {
			continue
		}

		newPath := filepath.Join(path, file.Name())

		if file.IsDir() {
			zones = walkTZDir(newPath, zones)
			continue
		}

		zones = append(zones, newPath)
	}

	return zones
}

func intRange(start, stop int) chan int {
	stream := make(chan int)

	if start <= stop {
		go func() {
			for i := start; i <= stop; i++ {
				stream <- i
			}
			close(stream)
		}()
	} else {
		go func() {
			for i := start; i >= stop; i-- {
				stream <- i
			}
			close(stream)
		}()
	}

	return stream
}

var functions = template.FuncMap{
	"pageControls":      ui.PageControls,
	"getOSTimeZones":    getOSTimeZones,
	"intRange":          intRange,
	"isoCountryToEmoji": isoCountryToEmoji,
	"stringsReplace":    strings.Replace,
}

type templateData struct {
	CSRFToken       string
	CurrentYear     int
	DarkMode        bool
	DiveSite        models.DiveSite
	DiveSites       []models.DiveSite
	Flash           string
	Form            any
	IsAuthenticated bool
	PageData        models.PageData
	WasPosted       bool
}

// https://stackoverflow.com/questions/26809484/how-to-use-double-star-glob-in-go
func recursiveGlob(filesystem fs.FS, dir string, ext string) ([]string, error) {
	files := []string{}

	err := fs.WalkDir(filesystem, dir, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ext {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pagesDir := "html/pages"
	pages := []string{}

	pages, err := recursiveGlob(ui.Files, pagesDir, ".tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := strings.TrimPrefix(page, pagesDir+"/")

		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		ts := template.New(name).Funcs(functions)
		ts, err = ts.ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
