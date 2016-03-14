// This package is a adapted from the "net/http" package
// by the The Go Authors. It has been adapted specifically
// for serving a static website with no directory browsing

// Original License:

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statichttp

import (
	"net/http"
	"os"
	"path"
	"strings"
)

type fileHandler struct {
	root http.FileSystem
}

// FileServer has been adapted from "net/http"
func FileServer(root http.FileSystem) http.Handler {
    //Quick change
	return &fileHandler{root}
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	serveFile(w, r, f.root, path.Clean(upath), true)
}

// name is '/'-separated, not filepath.Separator.
func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string, redirect bool) {
	const indexPage = "/index.html"

	// redirect .../index.html to .../
	// can't use Redirect() because that would make the path absolute,
	// which would be a problem running under StripPrefix
	if strings.HasSuffix(r.URL.Path, indexPage) {
		localRedirect(w, r, "./")
		return
	}

	f, err := fs.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err1 := f.Stat()
	if err1 != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	if redirect {
		// redirect to canonical path: / at end of directory url
		// r.URL.Path always begins with /
		url := r.URL.Path
		if d.IsDir() {
			if url[len(url)-1] != '/' {
				localRedirect(w, r, path.Base(url)+"/")
				return
			}
		} else {
			if url[len(url)-1] == '/' {
				localRedirect(w, r, "../"+path.Base(url))
				return
			}
		}
	}

	// use contents of index.html for directory, if present
	if d.IsDir() {
		index := strings.TrimSuffix(name, "/") + indexPage
		ff, err := fs.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				name = index
				d = dd
				f = ff
			}
		}
	}

	// Still a directory? (we didn't find an index.html file)
	if d.IsDir() {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// serveContent will check modification time
	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return http.StatusText(http.StatusNotFound), http.StatusNotFound
	}
	if os.IsPermission(err) {
		return http.StatusText(http.StatusForbidden), http.StatusForbidden
	}
	// Default:
	return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
}
