// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package main

import (
	"testing"
	"net/http/httptest"
	"net/http"
)

func TestLogin__forwardedPreflight(t *testing.T) {
	// setup test databases
	auth, err := createTestAuthable()
	if err != nil {
		t.Fatal(err)
	}
	defer auth.cleanup()

	repo, err := createTestUserRepository()
	if err != nil {
		t.Fatal(err)
	}
	defer repo.cleanup()

	// make our request
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users/login", nil)
	r.Header.Set("Origin", "http://localhost:8080")
	r.Header.Set("X-Forwarded-Method", "OPTIONS")

	handler := checkLogin(nil, auth, repo)
	handler(w, r)
	w.Flush()

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}
	if v := w.Header().Get("Access-Control-Allow-Origin"); v != "http://localhost:8080" {
		t.Errorf("got %s", v)
	}
}
