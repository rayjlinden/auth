// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

type testAuth struct {
	auth

	dir string
}

func (a *testAuth) cleanup() error {
	if err := a.auth.db.Close(); err != nil {
		return err
	}
	return os.RemoveAll(a.dir)
}

// createTestAuthable returns a new 'auth' instance
// wrapped with a cleanup() method.
//
// path is optional, if empty a new ioutil.TempDir will be
// created and optionally cleaned up.
func createTestAuthable() (*testAuth, error) {
	dir, err := ioutil.TempDir("", "auth")
	if err != nil {
		return nil, err
	}

	db, err := createConnection(filepath.Join(dir, "auth.db"))
	if err != nil {
		return nil, err
	}

	logger := log.NewLogfmtLogger(ioutil.Discard)
	if err := migrate(db, logger); err != nil {
		return nil, err
	}

	return &testAuth{auth{db, logger}, dir}, nil
}

type testUserRepository struct {
	sqliteUserRepository

	dir string
}

func (repo *testUserRepository) cleanup() error {
	if err := repo.sqliteUserRepository.close(); err != nil {
		return err
	}
	return os.RemoveAll(repo.dir)
}

// createTestUserRepository
func createTestUserRepository() (*testUserRepository, error) {
	dir, err := ioutil.TempDir("", "userRepository")
	if err != nil {
		return nil, err
	}

	db, err := createConnection(filepath.Join(dir, "auth.db"))
	if err != nil {
		return nil, err
	}

	logger := log.NewLogfmtLogger(ioutil.Discard)
	if err := migrate(db, logger); err != nil {
		return nil, err
	}

	return &testUserRepository{sqliteUserRepository{db, logger}, dir}, nil
}

func TestUser__cleanEmail(t *testing.T) {
	cases := []struct {
		input, expected string
	}{
		{"john.doe+moov@gmail.com", "johndoe@gmail.com"},
		{"john.doe+@gmail.com", "johndoe@gmail.com"},
		{"john.doe@gmail.com", "johndoe@gmail.com"},
		{"john.doe@gmail.com", "johndoe@gmail.com"},
		{"john+moov@gmail.com", "john@gmail.com"},
		{"john.@gmail.com", "john@gmail.com"},
		{"john.+@gmail.com", "john@gmail.com"},
	}

	for i := range cases {
		res := cleanEmail(cases[i].input)
		if res != cases[i].expected {
			t.Errorf("got %q", res)
		}
	}
}

func TestUserRepository(t *testing.T) {
	repo, err := createTestUserRepository()
	if err != nil {
		t.Fatal(err)
	}
	defer repo.cleanup()

	u, err := repo.lookupByUserId(generateID())
	if err != nil {
		t.Error(err)
	}
	if u != nil {
		t.Errorf("expected no user, got %v", u)
	}

	u, err = repo.lookupByEmail(generateID())
	if err != nil {
		t.Error(err)
	}
	if u != nil {
		t.Errorf("expected no user, got %v", u)
	}

	// create and insert user
	userId := generateID()
	u = &User{
		ID:        userId,
		Email:     "test@moov.io",
		FirstName: "Jane",
		LastName:  "Doe",
		Phone:     "111.222.3333",
		CreatedAt: time.Now().Add(-1 * time.Second),
	}

	if err := repo.upsert(u); err != nil {
		t.Fatal(err)
	}

	// make sure user was written
	uu, err := repo.lookupByUserId(userId)
	if err != nil {
		t.Error(err)
	}
	if uu == nil {
		t.Error("expected user")
	}

	uu, err = repo.lookupByEmail(u.Email)
	if err != nil {
		t.Error(err)
	}
	if uu == nil {
		t.Error("expected user")
	}

	// test upsert
	u.FirstName = "John"
	u.LastName = "Doe"
	u.Phone = "222.111.3333"
	u.CompanyURL = "https://moov.io"
	if err := repo.upsert(u); err != nil {
		t.Fatal(err)
	}

	// verify
	uu, err = repo.lookupByUserId(userId)
	if err != nil {
		t.Error(err)
	}
	if uu == nil {
		t.Error("expected user")
	}
	if u.ID != uu.ID {
		t.Errorf("u.ID=%q, uu.ID=%q", u.ID, uu.ID)
	}
	if u.FirstName != uu.FirstName {
		t.Errorf("u.FirstName=%q, uu.FirstName=%q", u.FirstName, uu.FirstName)
	}
	if u.LastName != uu.LastName {
		t.Errorf("u.LastName=%q, uu.LastName=%q", u.LastName, uu.LastName)
	}
	if u.Phone != uu.Phone {
		t.Errorf("u.Phone=%q, uu.Phone=%q", u.Phone, uu.Phone)
	}
	if u.CompanyURL != uu.CompanyURL {
		t.Errorf("u.CompanyURL=%q, uu.CompanyURL=%q", u.CompanyURL, uu.CompanyURL)
	}
	if !u.CreatedAt.Equal(uu.CreatedAt) {
		t.Errorf("u.CreatedAt=%q, uu.CreatedAt=%q", u.CreatedAt, uu.CreatedAt)
	}
}
