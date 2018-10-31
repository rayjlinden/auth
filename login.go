// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	moovhttp "github.com/moov-io/base/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func addLoginRoutes(router *mux.Router, logger log.Logger, auth authable, userService userRepository) {
	router.Methods("GET").Path("/users/login").HandlerFunc(checkLogin(logger, auth, userService))
	router.Methods("POST").Path("/users/login").HandlerFunc(loginRoute(logger, auth, userService))
}

func checkLogin(logger log.Logger, auth authable, userService userRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(w, r, "checkLogin")

		// Our LB setup uses "forward auth" which means traefik issues an internal
		// HTTP request to this method that checks each request for a valid cookie.
		// We use this instead of requiring each endpoint to be aware of auth.
		//
		// However, when a pre-flight request comes through that also triggers an
		// internal forward auth call into this method, but without the cookie.
		//
		// Thus, we need to check if the request was forwarded as a pre-flight request
		// and if so just respond with 200 and our usual CORS headers.
		origMethod := r.Header.Get("X-Forwarded-Method")
		if strings.EqualFold(origMethod, "OPTIONS") {
			moovhttp.SetAccessControlAllowHeaders(w, r)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Start checking the incoming request for cookie auth
		userId, err := extractUserId(auth, r)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		user, err := userService.lookupByUserId(userId)
		if err != nil {
			internalError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-User-Id", userId)
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(user); err != nil {
			internalError(w, err)
			return
		}
	}
}

func loginRoute(logger log.Logger, auth authable, userService userRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(w, r, "loginRoute")

		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bs, err := read(r.Body)
		if err != nil {
			internalError(w, err)
			return
		}

		// read request body
		var login loginRequest
		if err := json.Unmarshal(bs, &login); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Basic data sanity checks
		if err := validateEmail(login.Email); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := validatePassword(login.Password); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// find user by email
		u, err := userService.lookupByEmail(login.Email)
		if err != nil || u == nil {
			// Mark this (and password check) as failure only because
			// the user is involved at this point. Otherwise it's their
			// developer's problem (i.e. bad json).
			authFailures.With("method", "web").Add(1)
			w.WriteHeader(http.StatusForbidden)
			if err != nil {
				logger.Log("login", fmt.Sprintf("problem looking up user email %q: %v", login.Email, err))
			}
			return
		}

		// find user by userId and password
		if err := auth.checkPassword(u.ID, login.Password); err != nil {
			authFailures.With("method", "web").Add(1)
			logger.Log("login", fmt.Sprintf("userId=%s failed: %v", u.ID, err))
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// success route, let's finish!
		authSuccesses.With("method", "web").Add(1)
		cookie, err := createCookie(u.ID, auth)
		if err != nil {
			internalError(w, err)
			return
		}
		if cookie == nil {
			logger.Log("login", fmt.Sprintf("nil cookie for userId=%s", u.ID))
			internalError(w, err)
			return
		}
		if err := auth.writeCookie(u.ID, cookie); err != nil {
			internalError(w, err)
			return
		}

		http.SetCookie(w, cookie)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-User-Id", u.ID)
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(u); err != nil {
			internalError(w, err)
			return
		}
	}
}
