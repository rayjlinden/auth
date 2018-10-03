// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package buntdbclient

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/oauth2.v3/models"
)

var (
	flagDebug = flag.Bool("debug", false, "Create db inside project dir for tests")
)

type testCS struct {
	*ClientStore

	// temp dir used
	dir string
}

func (cs *testCS) cleanup() error {
	if cs == nil {
		return nil
	}
	err := cs.Close()
	if cs.dir != "" {
		os.RemoveAll(cs.dir)
	}
	return err
}

func makeCS(t *testing.T) (*testCS, error) {
	t.Helper()

	filename := "client_test.db"
	if *flagDebug {
		os.Remove(filename)
		db, err := New(filename)
		if err != nil {
			return nil, err
		}
		return &testCS{db, ""}, nil
	}

	dir, err := ioutil.TempDir("", "moov-auth-client")
	if err != nil {
		return nil, err
	}
	db, err := New(filepath.Join(dir, filename))
	if err != nil {
		return nil, err
	}
	return &testCS{db, dir}, nil
}

func TestClientStore(t *testing.T) {
	cs, err := makeCS(t)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.cleanup()

	id := "moov"

	// get nothing
	cli, err := cs.GetByID(id)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("got %#v", err)
	}
	if cli.GetID() != "" {
		t.Errorf("got %#v", err)
	}

	// set something
	err = cs.Set(id, &models.Client{
		ID:     id,
		Secret: "secret",
		Domain: "domain",
		UserID: "userId",
	})
	if err != nil {
		t.Errorf("got %v", err)
	}

	// get something
	cli, err = cs.GetByID(id)
	if err != nil {
		t.Errorf("got %v", err)
	}
	if cli.GetID() != id {
		t.Errorf("got %s", cli.GetID())
	}
	if cli.GetSecret() != "secret" {
		t.Errorf("got %s", cli.GetSecret())
	}
	if cli.GetDomain() != "domain" {
		t.Errorf("got %s", cli.GetDomain())
	}
	if cli.GetUserID() != "userId" {
		t.Errorf("got %s", cli.GetUserID())
	}
}

func TestClientStore__scan(t *testing.T) {
	cs, err := makeCS(t)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.cleanup()

	id, userId := "moov", "user-id"

	// scan nothing
	results, err := cs.GetByUserID(userId)
	if results != nil || err != nil {
		t.Errorf("got results=%v, err=%#v", results, err)
	}
	if v := len(results); v != 0 {
		t.Errorf("got %d", v)
	}

	// write something
	err = cs.Set(id, &models.Client{
		ID:     id,
		Secret: "secret",
		Domain: "domain",
		UserID: userId,
	})
	if err != nil {
		t.Errorf("got %v", err)
	}
	err = cs.Set(id+"2", &models.Client{
		ID:     id + "2",
		Secret: "secret",
		Domain: "domain",
		UserID: userId + "2",
	})
	if err != nil {
		t.Errorf("got %v", err)
	}
	err = cs.Set("other-id", &models.Client{
		ID:     "other-id",
		Secret: "secret",
		Domain: "domain",
		UserID: "other-user",
	})
	if err != nil {
		t.Errorf("got %v", err)
	}

	// scan something
	results, err = cs.GetByUserID(userId)
	if err != nil {
		t.Error(err)
	}
	if v := len(results); v != 1 {
		t.Errorf("got %d", v)
	}
}

func TestClientStore__delete(t *testing.T) {
	cs, err := makeCS(t)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.cleanup()

	id := "moov"

	// get nothing
	_, err = cs.GetByID(id)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("got %#v", err)
	}

	// set something
	cs.Set(id, &models.Client{
		ID:     id,
		Secret: "secret",
		Domain: "domain",
		UserID: "userId",
	})

	// get something
	cli, err := cs.GetByID(id)
	if err != nil || cli == nil {
		t.Errorf("got cli=%v, err=%#v", cli, err)
	}

	// delete
	cs.DeleteByID(id)

	// get nothing :-(
	_, err = cs.GetByID(id)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("got %#v", err)
	}
}
