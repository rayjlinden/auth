// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestHTTP__addCORS(t *testing.T) {
	router := mux.NewRouter()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "https://api.moov.io/v1/auth/ping", nil)
	r.Header.Set("Origin", "https://moov.io")

	addCORSHandler(router)
	router.ServeHTTP(w, r)
	w.Flush()

	if w.Code != 200 {
		t.Errorf("got %d", w.Code)
	}
	if v := w.Header().Get("Access-Control-Allow-Origin"); v != "*" {
		t.Errorf("got %q", v)
	}
	headers := []string{
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Credentials",
	}
	for i := range headers {
		v := w.Header().Get(headers[i])
		if v == "" {
			t.Errorf("%s: %q", headers[i], v)
		}
	}
}

// func TestHTTP__emptyOrigin(t *testing.T) {
// 	router := mux.NewRouter()
// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest("OPTIONS", "https://api.moov.io/v1/auth/ping", nil)
// 	r.Header.Set("Origin", "")

// 	addCORSHandler(router)
// 	router.ServeHTTP(w, r)
// 	w.Flush()

// 	if w.Code != http.StatusBadRequest {
// 		t.Errorf("got %d", w.Code)
// 	}
// }

func TestHTTP__extractCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	if req == nil {
		t.Error("nil req")
	}
	req.AddCookie(&http.Cookie{
		Name:  "moov_auth",
		Value: "data",
	})

	cookie := extractCookie(req)
	if cookie == nil {
		t.Error("nil cookie")
	}
}
