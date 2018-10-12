// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	errNoCookie     = fmt.Errorf("no %s cookie provided", cookieName)
	errUserNotFound = errors.New("user not found")
)

// extractUserId tries to find the Moov userId associated to the cookie data.
// This function will never return a non-blank userId with a non-nil error.
func extractUserId(auth authable, r *http.Request) (string, error) {
	cookie := extractCookie(r)
	if cookie == nil {
		return "", errNoCookie
	}
	userId, err := auth.findUserId(cookie.Value)
	if err != nil {
		return "", errUserNotFound
	}
	if userId == "" {
		return "", errUserNotFound
	}
	return userId, nil
}
