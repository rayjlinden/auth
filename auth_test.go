// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestAuth__extractUserId(t *testing.T) {
	auth, err := createTestAuthable()
	if err != nil {
		t.Fatal(err)
	}

	knownUserId := generateID()

	// Write our user
	cookie, err := createCookie(knownUserId, auth)
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.writeCookie(knownUserId, cookie); err != nil {
		t.Fatal(err)
	}

	// happy path
	req := httptest.NewRequest("GET", "/users/login", nil)
	req.Header.Set("cookie", fmt.Sprintf("moov_auth=%s", cookie.Value))

	// make our request
	userId, err := extractUserId(auth, req)
	if err != nil {
		t.Error(err)
	}
	if userId == "" {
		t.Error("empty userId")
	}
	if userId != knownUserId {
		t.Errorf("got %q, expected %q", userId, knownUserId)
	}

	// sad path
	req.Header.Set("cookie", "moov_auth=bad")
	userId, err = extractUserId(auth, req)
	if err == nil {
		t.Error("expected error")
	}
	if userId != "" {
		t.Errorf("got %s", userId)
	}
}
