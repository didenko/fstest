// Copyright 2017-2019 Vlad Didenko. All rights reserved.
// See the included LICENSE.md file for licensing information

package fst // import "go.didenko.com/fst"

import (
	"os"
	"path/filepath"
	"testing"
)

type DiffCase struct {
	dir   string
	comps []FileRank
}

func TestTreeDiff(t *testing.T) {

	_, cleanup := TempCloneChdir(t, "testdata/tree_diff_mocks")
	defer cleanup()

	FileDelAll(t, ".", "delete.me")

	successes := []DiffCase{
		{"a_same_content", []FileRank{ByName, ByDir, BySize, ByContent(t)}},
		{"d_same_empty", []FileRank{ByName, BySize}},
		{"e_same_empty_subdir", []FileRank{ByName, BySize}},
		{"k_same_size", []FileRank{ByName, BySize}},
		{"j_diff_sizes_same_perm", []FileRank{ByName, ByPerm}},
		{"l_perms_same", []FileRank{ByName, ByPerm}},
	}

	fails := []DiffCase{
		{"b_left_nodir", []FileRank{ByName}},
		{"b_right_nodir", []FileRank{ByName}},
		{"c_left_nofile", []FileRank{ByName}},
		{"c_right_nofile", []FileRank{ByName}},
		{"f_dir_left_file_right", []FileRank{ByName, ByDir}},
		{"f_dir_right_file_left", []FileRank{ByName, ByDir}},
		{"g_empty_left", []FileRank{ByName}},
		{"g_empty_right", []FileRank{ByName}},
		{"h_diff_content_bin", []FileRank{ByName, ByContent(t)}},
		{"i_diff_content_text_eol", []FileRank{ByName, ByContent(t)}},
		{"j_diff_sizes_same_perm", []FileRank{ByName, BySize}},
		{"l_perms_same", []FileRank{ByName, ByPerm, BySize}},
	}

	for _, tc := range successes {

		diffs := TreeDiff(t, filepath.Join(tc.dir, "a"), filepath.Join(tc.dir, "b"), tc.comps...)

		if diffs != nil {
			t.Errorf("Equivalent directories in \"%s\" tested as different: %v\n", tc.dir, diffs)
		}
	}

	for _, tc := range fails {

		diffs := TreeDiff(t, filepath.Join(tc.dir, "a"), filepath.Join(tc.dir, "b"), tc.comps...)

		if diffs == nil {
			t.Errorf("Differing directories in \"%s\" passed as equivalent\n", tc.dir)
		}
	}
}

// Time comparisons are tested separately because Git doe not
// preserve timestamps - meaning it is impossible to enforce
// timestamps of checked out files and directories without
// jumping through extra hoops (possibly, git hooks would do)
//
// Here it is simpler to use the TempCreateChdir function which
// sets timestamps correctly instead of fighting git.
func TestTreeDiffTimes(t *testing.T) {

	mocks, err := os.Open("testdata/tree_diff_time_mocks")
	if err != nil {
		t.Fatal(err)
	}

	nodes := ParseReader(t, mocks)

	_, cleanup := TempCreateChdir(t, nodes)
	defer cleanup()

	diffs := TreeDiff(t, "a_same_times/a", "a_same_times/b", ByName, ByTime)

	if diffs != nil {
		t.Errorf("Equivalent directories in \"%s\" tested as different: %v\n", "a_same_times", diffs)
	}

	diffs = TreeDiff(t, "b_diff_time_file/a", "b_diff_time_file/b", ByName, ByTime)

	if diffs == nil {
		t.Errorf("Differing directories in \"%s\" passed as equivalent\n", "b_diff_time_file")
	}

	diffs = TreeDiff(t, "c_diff_time_dir/a", "c_diff_time_dir/b", ByName, ByTime)

	if diffs == nil {
		t.Errorf("Differing directories in \"%s\" passed as equivalent\n", "c_diff_time_dir")
	}
}
