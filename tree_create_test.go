// Copyright 2017 Vlad Didenko. All rights reserved.
// See the included LICENSE.md file for licensing information

package fst // import "go.didenko.com/fst"

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

type tcase struct {
	t time.Time
	p os.FileMode
	n string
	c string
}

func match(t *testing.T, f *tcase, fi os.FileInfo) {

	if !fi.ModTime().Equal(f.t) {
		t.Errorf("Times mismatch, expected \"%v\", got \"%v\" for \"%v\"", f.t, fi.ModTime().UTC(), f.n)
	}
	if fi.Mode().Perm() != f.p {
		t.Errorf("Permissions mismatch, expected \"%v\", got \"%v\" for \"%v\"", f.p, fi.Mode().Perm(), f.n)
	}

	if f.c == "" {
		return
	}

	byteContent, err := ioutil.ReadFile(f.n)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal([]byte(f.c), byteContent) {
		t.Errorf("Content mismatch, expected \"%v\", got \"%v\" for \"%v\"", f.c, byteContent, f.n)
	}
}

func TestTreeCreateFromReader(t *testing.T) {

	dirs := `
		2001-01-01T01:01:01Z	0150	aaa/
		2099-01-01T01:01:01Z	0700	aaa/bbb/

		2001-01-01T01:01:01Z	0700	c.txt	"This is a two line\nfile with\ta tab\n"
		2001-01-01T01:01:01Z	0700	d.txt	No need to quote a single line without tabs

		2002-01-01T01:01:01Z	0700	"has\ttab/"
		2002-01-01T01:01:01Z	0700	"has\ttab/e.mb"	"# Markdown...\n\n... also ***possible***\n"

		2002-01-01T01:01:01Z	0700	"\u10077heavy quoted\u10078/"`

	expect := []tcase{
		{time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), 0150, "aaa", ""},
		{time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "c.txt", ""},
		{time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "d.txt", "No need to quote a single line without tabs"},
		{time.Date(2002, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "has\ttab", ""},
		{time.Date(2002, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "\u10077heavy quoted\u10078", ""},
		{time.Date(2099, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "aaa/bbb", ""},
	}

	_, cleanup, err := TempInitChdir()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	nodes, err := TreeParseReader(strings.NewReader(dirs))
	if err != nil {
		t.Fatal(err)
	}

	err = TreeCreate(nodes)
	if err != nil {
		t.Fatal(err)
	}

	files, err := ioutil.ReadDir(".")
	if err != nil {
		t.Fatal(err)
		return
	}

	for i, fi := range files {
		if fi.Name() != expect[i].n {
			t.Errorf("Names mismatch, expected \"%v\", got \"%v\"", expect[i].n, fi.Name())
		}
		match(t, &expect[i], fi)
	}

	for _, tc := range expect {
		f := tc.n
		fi, err := os.Stat(f)
		if err != nil {
			t.Fatal(err)
		}

		match(t, &tc, fi)
	}
}

func TestTreeCreateApi(t *testing.T) {
	items := []*Node{
		&Node{name: "aaa/", perm: 0150, time: time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC)},
		&Node{name: "c.txt", perm: 0700, time: time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), body: "This is a two line\nfile with\ta tab\n"},
		&Node{name: "d.txt", perm: 0700, time: time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), body: "No need to quote a single line without tabs"},
		&Node{name: "has\ttab/", perm: 0700, time: time.Date(2002, time.January, 1, 1, 1, 1, 0, time.UTC)},
		&Node{name: "has\ttab/e.mb", perm: 0700, time: time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), body: ""},
		&Node{name: "\u10077heavy quoted\u10078", perm: 0700, time: time.Date(2002, time.January, 1, 1, 1, 1, 0, time.UTC), body: ""},
		&Node{name: "aaa/bbb", perm: 0700, time: time.Date(2099, time.January, 1, 1, 1, 1, 0, time.UTC), body: ""},
	}

	expect := []tcase{
		{time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), 0150, "aaa", ""},
		{time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "c.txt", ""},
		{time.Date(2001, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "d.txt", "No need to quote a single line without tabs"},
		{time.Date(2002, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "has\ttab", ""},
		{time.Date(2002, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "\u10077heavy quoted\u10078", ""},
		{time.Date(2099, time.January, 1, 1, 1, 1, 0, time.UTC), 0700, "aaa/bbb", ""},
	}

	_, cleanup, err := TempInitChdir()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	err = TreeCreate(items)
	if err != nil {
		t.Fatal(err)
	}

	files, err := ioutil.ReadDir(".")
	if err != nil {
		t.Fatal(err)
		return
	}

	for i, fi := range files {
		if fi.Name() != expect[i].n {
			t.Errorf("Names mismatch, expected \"%v\", got \"%v\"", expect[i].n, fi.Name())
		}
		match(t, &expect[i], fi)
	}

	for _, tc := range expect {
		f := tc.n
		fi, err := os.Stat(f)
		if err != nil {
			t.Fatal(err)
		}

		match(t, &tc, fi)
	}
}
