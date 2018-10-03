// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"testing"
)

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
