// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
	"testing"
)

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
