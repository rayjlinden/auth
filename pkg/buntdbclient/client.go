// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// buntdbclient implements ClientStore from gopkg.in/oauth2.v3
// using BuntDB (https://github.com/tidwall/buntdb).
//
// A few extra operations have been added though, such as Set and
// GetByUserId. These were needed for ourusecase as we're mutating
// the oauth clients.
//
// Tests can be ran with a database in the package dir, just add -debug
// as a flag to 'go test'.
// The local database will be deleted before tests are ran each time.
package buntdbclient

import (
	"fmt"

	"github.com/tidwall/buntdb"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

// New initializes a new BuntDB database with any indicies needed
// for the Get* operations.
func New(path string) (*ClientStore, error) {
	db, err := buntdb.Open(path)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *buntdb.Tx) error {
		return tx.CreateIndex("user_id", "*", buntdb.IndexJSON("UserID")) // ScanByUserId
	})
	if err != nil {
		return nil, fmt.Errorf("problem running migrations: %v", err)
	}

	return &ClientStore{
		db: db,
	}, nil
}

// ClientStore wraps oauth2.ClientStore
type ClientStore struct {
	oauth2.ClientStore

	db *buntdb.DB
}

// Close shuts down connections to the underlying database
func (cs *ClientStore) Close() error {
	return cs.db.Close()
}

// GetById returns an oauth2.ClientInfo if the ID matches id.
func (cs *ClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	var cli models.Client
	cli.ID = id

	err := cs.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(fmt.Sprintf("%s-secret", id))
		if err != nil {
			return err
		}
		cli.Secret = v

		v, err = tx.Get(fmt.Sprintf("%s-domain", id))
		if err != nil {
			return err
		}
		cli.Domain = v

		v, err = tx.Get(fmt.Sprintf("%s-user-id", id))
		if err != nil {
			return err
		}
		cli.UserID = v
		return nil
	})
	if err != nil {
		var cli models.Client
		return &cli, fmt.Errorf("problem reading %s: %v", id, err)
	}
	return &cli, nil
}

// Set writes the oauth2.ClientInfo to the underlying database.
func (cs *ClientStore) Set(id string, cli oauth2.ClientInfo) error {
	if inc := cli.GetID(); id != inc {
		return fmt.Errorf("ClientStore: id's don't match, id=%s and cli=%s", id, inc)
	}

	err := cs.db.Update(func(tx *buntdb.Tx) error {
		opts := &buntdb.SetOptions{}
		_, _, err := tx.Set(fmt.Sprintf("%s-secret", id), cli.GetSecret(), opts)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(fmt.Sprintf("%s-domain", id), cli.GetDomain(), opts)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(fmt.Sprintf("%s-user-id", id), cli.GetUserID(), opts)
		return err
	})
	if err != nil {
		return fmt.Errorf("problem updating %s: %v", id, err)
	}
	return nil
}

// GetByUserID returns an array of oauth2.ClientInfo which have a UserID mathcing
// userId.
// If return values are nil that means no matching records were found.
func (cs *ClientStore) GetByUserID(userId string) ([]oauth2.ClientInfo, error) {
	keys := make(map[string]bool)

	err := cs.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual("user_id", userId, func(k, v string) bool {
			if v == userId {
				keys[k] = true
			}
			if k > userId {
				// quit iterating once we go "past" userId
				return false
			}
			return true
		})
	})
	if err != nil {
		return nil, err
	}

	// Grab each ClientInfo now
	var accum []oauth2.ClientInfo
	for k := range keys {
		ci, _ := cs.GetByID(k)
		if ci != nil {
			accum = append(accum, ci)
		}
	}
	if len(accum) == 0 {
		return nil, nil
	}
	return accum, nil
}

// DeleteByID removes the oauth2.ClientInfo for the provided id.
func (cs *ClientStore) DeleteByID(id string) error {
	return cs.db.Update(func(tx *buntdb.Tx) (e error) {
		tx.Delete(fmt.Sprintf("%s-secret", id))
		tx.Delete(fmt.Sprintf("%s-domain", id))
		_, err := tx.Delete(fmt.Sprintf("%s-user-id", id))
		return err
	})
}
