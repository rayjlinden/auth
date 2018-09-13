// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/moov-io/auth/pkg/buntdbclient"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

type oauth struct {
	manager     *manage.Manager
	clientStore *buntdbclient.ClientStore
	server      *server.Server

	logger log.Logger
}

func setupOauthServer(logger log.Logger) (*oauth, error) {
	out := &oauth{
		logger: logger,
	}

	// oauth2 setup
	path := os.Getenv("OAUTH2_TOKENS_DB_PATH")
	if path == "" {
		path = "oauth2_tokens.db"
	}
	tokenStore, err := store.NewFileTokenStore(path)
	if err != nil {
		return nil, fmt.Errorf("problem creating token store: %v", err)
	}

	out.manager = manage.NewDefaultManager()
	out.manager.MapTokenStorage(tokenStore)

	path = os.Getenv("OAUTH2_CLIENTS_DB_PATH")
	if path == "" {
		path = "oauth2_clients.db"
	}
	cs, err := buntdbclient.New(path)
	if err != nil {
		return nil, fmt.Errorf("problem creating clients store: %v", err)
	}
	out.clientStore = cs
	out.manager.MapClientStorage(out.clientStore)

	out.server = server.NewDefaultServer(out.manager)
	out.server.SetAllowGetAccessRequest(true)
	out.server.SetClientInfoHandler(server.ClientFormHandler)
	out.server.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		logger.Log("internal-error", err.Error())
		return
	})
	out.server.SetResponseErrorHandler(func(re *errors.Response) {
		logger.Log("response-error", re.Error.Error())
		return
	})

	return out, nil
}

// addOAuthRoutes includes our oauth2 routes on the provided mux.Router
func addOAuthRoutes(r *mux.Router, o *oauth, logger log.Logger, auth authable) {
	r.Methods("GET").Path("/authorize").HandlerFunc(o.authorizeHandler)
	if o.server.Config.AllowGetAccessRequest {
		r.Methods("GET").Path("/token").HandlerFunc(o.tokenHandler)
	} else {
		r.Methods("POST").Path("/token").HandlerFunc(o.tokenHandler)
	}
	r.Methods("POST").Path("/token/recreate").HandlerFunc(o.recreateTokenHandler(auth))
}

// authorizeHandler checks the request for appropriate oauth information
// and returns "200 OK" if the token is valid.
func (o *oauth) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	// We aren't using HandleAuthorizeRequest here because that assumes redirect_uri
	// exists on the request. We're just checking for a valid token.
	ti, err := o.server.ValidationBearerToken(r)
	if err != nil {
		authFailures.With("method", "oauth2").Add(1)
		encodeError(w, err)
		return
	}
	if ti.GetClientID() == "" {
		authFailures.With("method", "oauth2").Add(1)
		encodeError(w, fmt.Errorf("missing client_id"))
		return
	}

	// Passed token check, return "200 OK"
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	authSuccesses.With("method", "oauth2").Add(1)
}

// tokenHandler passes off the request down to our oauth2 library to
// generate a token (or return an error).e
func (o *oauth) tokenHandler(w http.ResponseWriter, r *http.Request) {
	err := o.server.HandleTokenRequest(w, r)
	if err != nil {
		encodeError(w, err)
		return
	}
	// TODO(adam): We need to track this metric inside our TokenStorage.
	// HandleTokenRequest currently returns nil even if the token request
	// failed. There's no real way to inspect the http.ResponseWriter in
	// an attempt to correctly calculate this.
	// tokenGenerations.With("method", "oauth2").Add(1)
}

// recreateTokenHandler will recreate the oauth token for a user. This involves:
//  - invalidate all existing tokens
//  - creates new tokens (and returns them only once)
//
// This method extracts the user from the cookies in r.
func (o *oauth) recreateTokenHandler(auth authable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := auth.findUserId(extractCookie(r).Value)
		if err != nil {
			internalError(w, err, "oauth")
			return
		}

		records, err := o.clientStore.GetByUserID(userId)
		if records == nil && err == nil { // nothing found
			o.logger.Log("oauth", fmt.Sprintf("userId=%s had no oauth clients", userId))
			w.WriteHeader(http.StatusOK)
			return
		}

		clients := make([]*models.Client, len(records))
		for i := range records {
			if err := o.clientStore.DeleteByID(records[i].GetID()); err != nil {
				internalError(w, err, "oauth")
				return
			}

			clients[i] = &models.Client{
				ID:     generateID()[:12],
				Secret: generateID(),
				Domain: records[i].GetDomain(),
				UserID: userId,
			}

			if err := o.clientStore.Set(records[i].GetID(), records[i]); err != nil {
				internalError(w, err, "oauth")
				return
			}
		}

		// metrics
		tokenGenerations.With("method", "oauth2_via_web").Add(1)

		// render back new client info
		type response struct {
			Clients []*models.Client `json:"clients"`
		}
		if err := json.NewEncoder(w).Encode(&response{clients}); err != nil {
			internalError(w, err, "oauth")
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
}

func (o *oauth) shutdown() error {
	if o == nil || o.clientStore == nil {
		return nil
	}
	return o.clientStore.Close()
}
