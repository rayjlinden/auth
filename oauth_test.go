// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
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

type testOAuth struct {
	svc *oauth
	dir string

	tokenStore oauth2.TokenStore
}

func (o *testOAuth) cleanup() error {
	defer os.RemoveAll(o.dir)
	return o.svc.shutdown()
}

func createTestOAuth() (*testOAuth, error) {
	dir, err := ioutil.TempDir("", "auth-oauth")
	if err != nil {
		return nil, err
	}

	tokenStore, err := setupOAuthTokenStore(filepath.Join(dir, "oauth2_tokens.db"))
	if err != nil {
		return nil, err
	}
	clientStore, err := setupOAuthClientStore(filepath.Join(dir, "oauth2_clients.db"))
	if err != nil {
		return nil, err
	}

	svc, err := setupOAuthServer(log.NewNopLogger(), tokenStore, clientStore)
	if err != nil {
		return nil, err
	}

	return &testOAuth{
		svc: svc,
		dir: dir,

		tokenStore: tokenStore,
	}, nil
}

func createOAuthToken(t *testing.T, o *testOAuth) *models.Token {
	t.Helper()

	token := &models.Token{
		ClientID:        generateID(),
		Scope:           "read",
		Access:          generateID(),
		AccessCreateAt:  time.Now().Add(-1 * time.Second), // in the past
		AccessExpiresIn: 30 * time.Minute,                 // the future
	}
	if err := o.tokenStore.Create(token); err != nil {
		t.Fatal(err)
		return nil
	}
	return token
}

func TestOAuth__BearerToken(t *testing.T) {
	o, err := createTestOAuth()
	if err != nil {
		t.Fatal(err)
	}
	defer o.cleanup()

	// missing header
	req := httptest.NewRequest("GET", "/users/login", nil)
	if err := o.svc.requestHasValidOAuthToken(req); err == nil {
		t.Errorf("expected error, no Authorization header set")
	}

	// bad header value
	req.Header.Set("Authorization", "Bearer bad")
	if err := o.svc.requestHasValidOAuthToken(req); err == nil {
		t.Errorf("expected error, no bad Authorization data")
	}

	// happy path
	token := createOAuthToken(t, o)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.GetAccess()))
	if err := o.svc.requestHasValidOAuthToken(req); err != nil {
		t.Errorf("expected no error: %v", err)
	}
}
