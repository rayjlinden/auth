// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	// maxReadBytes is the number of bytes to read
	// from a request body. It's intended to be used
	// with an io.LimitReader
	maxReadBytes = 1 * 1024 * 1024

	cookieName = "moov_auth"
	cookieTTL  = 30 * 24 * time.Hour // days * hours/day * hours
)

var (
	// Domain is the domain to publish cookies under.
	// If empty "localhost" is used.
	//
	// The path is always set to /.
	Domain string = os.Getenv("DOMAIN")
)

func init() {
	if Domain == "" {
		Domain = "localhost"
	}
}

// read consumes an io.Reader (wrapping with io.LimitReader)
// and returns either the resulting bytes or a non-nil error.
func read(r io.Reader) ([]byte, error) {
	r = io.LimitReader(r, maxReadBytes)
	return ioutil.ReadAll(r)
}

// encodeError JSON encodes the supplied error
//
// The HTTP status of "400 Bad Request" is written to the
// response.
func encodeError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func internalError(w http.ResponseWriter, err error, component string) {
	internalServerErrors.Add(1)
	if logger != nil {
		logger.Log(component, err)
	}
	w.WriteHeader(http.StatusInternalServerError)
}

// extractCookie attempts to pull out our cookie from the incoming request.
// We use the contents to find the associated userId.
func extractCookie(r *http.Request) *http.Cookie {
	if r == nil {
		return nil
	}
	cs := r.Cookies()
	for i := range cs {
		if cs[i].Name == cookieName {
			return cs[i]
		}
	}
	return nil
}

// createCookie generates a new cookie and associates it with the provided
// userId.
func createCookie(userId string, auth authable) (*http.Cookie, error) {
	cookie := &http.Cookie{
		Domain:   Domain,
		Expires:  time.Now().Add(cookieTTL),
		HttpOnly: true,
		Name:     cookieName,
		Path:     "/",
		Secure:   serveViaTLS,
		Value:    generateID(),
	}
	if err := auth.writeCookie(userId, cookie); err != nil {
		return nil, err
	}
	return cookie, nil
}

// addCORSHandler captures Corss Origin Resource Sharing (CORS) requests
// by looking at all OPTIONS requests for the Origin header, parsing that
// and responding back with the other Access-Control-Allow-* headers.
//
// Docs: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
func addCORSHandler(r *mux.Router) {
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			if logger != nil {
				line := fmt.Sprintf("method=%s, path=%s, preflight - no origin", r.Method, r.URL.Path)
				logger.Log("http", line)
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		setAccessControlAllow(w, r)
		w.WriteHeader(http.StatusOK)
	})
}

func setAccessControlAllow(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	// Access-Control-Allow-Origin can't be '*' with requests that send credentials.
	// Instead, we need to explicitly set the domain (from request's Origin header)
	//
	// Allow requests from anyone's localhost and only from secure pages.
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "https://") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "PATCH,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Cookie,X-User-Id,X-Request-Id,Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Content-Type", "text/plain")
	}
}

// getRequestId extracts X-Request-Id from the http request, which
// is used in tracing requests.
//
// TODO(adam): IIRC a "max header size" param in net/http.Server - verify and configure
func getRequestId(r *http.Request) string {
	return r.Header.Get("X-Request-Id")
}

func addPingRoute(r *mux.Router) {
	r.Methods("GET").Path("/ping").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(w, r, "ping")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PONG"))
	})
}

// responsewriter is an http.ResponseWriter that records the
// outgoing response and executes a callback with the saved
// metadata.
type responseWriter struct {
	rec *httptest.ResponseRecorder
	w   http.ResponseWriter

	start          time.Time
	method         string
	request        *http.Request
	headersWritten bool
}

func (w *responseWriter) requestId() string {
	return getRequestId(w.request)
}

// Header returns headers added to the response
func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

// Write copies the sent bytes and records them, but then
// sends the incoming array down to our empty http.ResponseWriter
func (w *responseWriter) Write(b []byte) (int, error) {
	bb := make([]byte, len(b))
	if copy(bb, b) == 0 {
		return 0, errors.New("empty body")
	}
	w.rec.Write(bb)
	return w.w.Write(b)
}

// WriteHeader sets the status code for the request and flushes headers
// back to the client.
//
// The provided callback is executed after flushing headers.
func (w *responseWriter) WriteHeader(code int) {
	if w.headersWritten {
		return
	}
	w.headersWritten = true

	setAccessControlAllow(w, w.request)

	w.rec.WriteHeader(code)
	w.w.WriteHeader(code)

	w.callback()
}

func (w *responseWriter) callback() {
	diff := time.Now().Sub(w.start)
	routeHistogram.With("route", w.method).Observe(diff.Seconds())

	if w.method != "" && w.requestId() != "" {
		line := strings.Join([]string{
			fmt.Sprintf("method=%s", w.request.Method),
			fmt.Sprintf("path=%s", w.request.URL.Path),
			fmt.Sprintf("status=%d", w.rec.Code),
			fmt.Sprintf("took=%s", diff),
			fmt.Sprintf("requestId=%s", w.requestId()),
		}, ", ")
		logger.Log(w.method, line)
	}
}

// wrapResponseWriter creates a new responseWriter and initializes an
// httptest.ResponseRecorder
func wrapResponseWriter(w http.ResponseWriter, r *http.Request, method string) http.ResponseWriter {
	return &responseWriter{
		method:  method,
		request: r,
		rec:     httptest.NewRecorder(),
		w:       w,
		start:   time.Now(),
	}
}
