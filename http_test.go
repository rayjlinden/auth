// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if v := w.Header().Get("Access-Control-Allow-Origin"); v != "https://moov.io" {
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

func TestHTTP__emptyOrigin(t *testing.T) {
	router := mux.NewRouter()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "https://api.moov.io/v1/auth/ping", nil)
	r.Header.Set("Origin", "")

	addCORSHandler(router)
	router.ServeHTTP(w, r)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
}

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

func TestHTTP_createCookie(t *testing.T) {
	auth, err := createTestAuthable()
	if err != nil {
		t.Fatal(err)
	}
	defer auth.cleanup()

	userId := generateID()
	cookie, err := createCookie(userId, auth)
	if err != nil {
		t.Fatal(err)
	}

	if id, _ := auth.findUserId(cookie.Value); id != userId {
		t.Errorf("got %q, expected %q", id, userId)
	}
}

func TestHTTP_encodeError(t *testing.T) {
	w := httptest.NewRecorder()
	encodeError(w, errors.New("problem X"))
	w.Flush()

	// check http response
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
	v := w.Header().Get("Content-Type")
	if !strings.Contains(v, "application/json") {
		t.Errorf("got %s", v)
	}

	type resp struct {
		Error string `json:"error"`
	}
	var response resp
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Error(err)
	}
	if response.Error != "problem X" {
		t.Errorf("got %q", response.Error)
	}
}

func TestHTTP_internalError(t *testing.T) {
	w := httptest.NewRecorder()
	internalError(w, errors.New("problem Y"), "test")
	w.Flush()

	if w.Code != http.StatusInternalServerError {
		t.Errorf("got %d", w.Code)
	}
}
