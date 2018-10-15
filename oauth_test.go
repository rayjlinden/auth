// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"testing"
)

func TestOAuth2_clientsJSON(t *testing.T) {
	raw := []byte(`[{"client_id": "foo", "client_secret": "secrets", "domain": "moov.io"}]`)

	var clients []*client
	if err := json.Unmarshal(raw, &clients); err != nil {
		t.Fatal(err)
	}

	if len(clients) != 1 {
		t.Errorf("got %#v", clients)
	}

	if clients[0].ClientID != "foo" {
		t.Errorf("got %s", clients[0].ClientID)
	}
	if clients[0].ClientSecret != "secrets" {
		t.Errorf("got %s", clients[0].ClientSecret)
	}
	if clients[0].Domain != "moov.io" {
		t.Errorf("got %s", clients[0].Domain)
	}
}
