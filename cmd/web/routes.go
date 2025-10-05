package main

import (
	"net/http"
	"path/filepath"

	"github.com/justinas/alice"
	"github.com/m5lapp/divesite-monolith/ui"
)

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()

	fileserver := http.FileServer(neuteredFileSystem{fs: http.FS(ui.Files)})
	mux.Handle("GET /static", http.NotFoundHandler())
	mux.Handle("GET /static/", fileserver)

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.HandleFunc("GET /status", status)

	mux.Handle("GET  /user/sign-up", dynamic.ThenFunc(app.userCreateGET))
	mux.Handle("POST /user/sign-up", dynamic.ThenFunc(app.userCreatePOST))
	mux.Handle("GET  /user/log-in", dynamic.ThenFunc(app.userLogInGET))
	mux.Handle("POST /user/log-in", dynamic.ThenFunc(app.userLogInPOST))
	mux.Handle("POST /user/log-out", protected.ThenFunc(app.userLogOutPOST))

	mux.Handle("GET  /log-book/dive/", protected.ThenFunc(app.diveList))
	mux.Handle("GET  /log-book/dive/add", protected.ThenFunc(app.diveCreateGET))
	mux.Handle("POST /log-book/dive/add", protected.ThenFunc(app.diveCreatePOST))
	mux.Handle("GET  /log-book/dive/view/{id}", protected.ThenFunc(app.diveGET))
	mux.Handle("GET  /log-book/dive-site/add", protected.ThenFunc(app.diveSiteCreateGET))
	mux.Handle("POST /log-book/dive-site/add", protected.ThenFunc(app.diveSiteCreatePOST))
	mux.Handle("GET  /log-book/dive-site", protected.ThenFunc(app.diveSiteList))
	mux.Handle("GET  /log-book/dive-site/view/{id}", protected.ThenFunc(app.diveSiteGET))

	mux.Handle("GET  /buddy/", protected.ThenFunc(app.buddyList))
	mux.Handle("GET  /buddy/add", protected.ThenFunc(app.buddyCreateGET))
	mux.Handle("POST /buddy/add", protected.ThenFunc(app.buddyCreatePOST))

	mux.Handle("GET  /certification/", protected.ThenFunc(app.certificationList))
	mux.Handle("GET  /certification/add", protected.ThenFunc(app.certificationCreateGET))
	mux.Handle("POST /certification/add", protected.ThenFunc(app.certificationCreatePOST))

	mux.Handle("GET  /operator/", protected.ThenFunc(app.operatorList))
	mux.Handle("GET  /operator/add", protected.ThenFunc(app.operatorCreateGET))
	mux.Handle("POST /operator/add", protected.ThenFunc(app.operatorCreatePOST))

	mux.Handle("GET  /trip/", protected.ThenFunc(app.tripList))
	mux.Handle("GET  /trip/add", protected.ThenFunc(app.tripCreateGET))
	mux.Handle("POST /trip/add", protected.ThenFunc(app.tripCreatePOST))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		index := filepath.Join("path", "index.html")
		_, err := nfs.fs.Open(index)
		if err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
