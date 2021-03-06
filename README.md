# **fst** - File System Testing aids

[![GoDoc](https://godoc.org/go.didenko.com/fst?status.svg)](https://godoc.org/go.didenko.com/fst)
[![Build Status](https://travis-ci.org/didenko/fst.svg?branch=master)](https://travis-ci.org/didenko/fst)
[![Go Report Card](https://goreportcard.com/badge/go.didenko.com/fst)](https://goreportcard.com/report/go.didenko.com/fst)

The suggested package name pronounciation is _"fist"_.

## Purpose

Sometimes it is desireable to test a program behavior which creates or modifies files and directories. Such tests may be quite involved especially if checking permissions or timestamps. A proper cleanup is also considered a nuisance. The whole effort becomes extra burdesome as such filesystem manipulation has to be tested itself - so one ends up with tests of tests.

The `fst` library is a tested set of functions aiming to alleviate the burden. It makes creating and comparing filesystem trees of regular files and directories for testing puposes simple.

## Highlights

The three most used functtions in the `fst` library are [_TempCloneChdir_](#TempCloneChdir), [_TempCreateChdir_](#TempCreateChdir), and [_TreeDiff_](#TreeDiff). For details on these and other functions, see the examples and documentation at the https://godoc.org/go.didenko.com/fst page.

### <span id="TempCloneChdir" />[TempCloneChdir](https://godoc.org/go.didenko.com/fst#TempCloneDir)

TempCloneChdir is intended to clone an existing directory with all its content, permissions, and timestamps. Consider this example:

```go
old, cleanup := fst.TempCloneChdir(t, "mock")
defer cleanup()
```

If the operation failed then (a) no temporary directory was left behind after a reasonable cleanup effort and (b) the `t.Fatalf(...)` was called.

If the operation succeded then the _mock_ directory's content is be cloned into a new temporary directory and the calling process will change working directory into it. The _old_ variable in the example will contain the original directory where the running process was at the time of the _TempCloneChdir_ call.

The ***cleanup*** function has the code to change back to the original directory and then delete the temporary directory.

As the _TempCloneChdir_ relies on the `TreeCopy` function, it will attempt to recreate both permissions and timestamps from the source directory. Keep in mind, that popular version control systems like _Git_ and _Mercurial_ do not preserve original files' timestamps. If your tests rely on timestamped files or directories then _TreeCreate_ or its derivative _TempCreateChdir_ functions are your friends.

### <span id="TempCreateChdir" />[TempCreateChdir](https://godoc.org/go.didenko.com/fst#TempCreateChdir)

The _TempCreateChdir_ function provides an API-like way to create and populate a temporary directory tree for testing. It takes a slice of `Node` pointers, from which it expects to receive data describing the directories and files to be populated. Here is an a hypothetical example:

```go
nodes := []*fst.Node{
  &fst.Node{0750, fst.Rfc3339(t, "2017-11-12T13:14:15Z"), "settings/", ""},
  &fst.Node{0640, fst.Rfc3339(t, "2017-11-12T13:14:15Z"), "settings/theme1.toml", "key = val1"},
  &fst.Node{0640, fst.Rfc3339(t, "2017-11-12T13:14:15Z"), "settings/theme2.toml", "key = val2"},
}

old, cleanup := fst.TempCreateChdir(t, nodes)
defer cleanup()
```

If there was no failure then _TempCreateChdir_ will create the `settings` directory  with files `theme1.toml` and `theme2.toml`, with the specified key/value pairs as content in the `settings` directory.

As a part of the cleanup logic the _TempCreateChdir_ function removes the temporary directory it created. It also does a best-effort attempt to remove the temporary directory if it failed during its operation, while also calling `t.Fatalf(...)`.

### <span id="TreeDiff" />[_TreeDiff_](https://godoc.org/go.didenko.com/fst#TreeDiff)

The _TreeDiff_ function produces a human-readable output of differences between two directory trees for diagnostic purposes. The resulting slice of strings is empty if no differences are found.

Criteria for comparing filesystem objects varies based on a task, so _TreeDiff_ takes a list of comparator functions. The most common ones are provided with the `fst` package. Users are free to provide their own additional comparators which satisfy the [_FileRank_](https://godoc.org/go.didenko.com/fst#FileRank) signature.

A quick example of a common _TreeDiff_ use:

```go
diffs:= fst.TreeDiff(
  t,
  "dir1", "dir2",
  fst.ByName, fst.ByDir, fst.BySize, fst.ByContent(t))

if diffs != nil {
  t.Logf("Differences between dir1 and dir2:\n%v\n", diffs)
}
```

Note, that while the _BySize_ comparator is redundant in presense of the _ByContent_ comparator, in most cases the cheaper size comparison will avoid a more expensive content comparison. The comparator order is significant, because once an earlier comparator returns `true`, the later comparators do not run.

It is easy to provide overly restrictive permissions using the tree cloning and tree creation functions. When unable to access needed information, _TreeDiff_ will call `t.Fatalf(...)` with a related diagnostic. While specifics may vary it is often safest to set user read and execute permissions for directories and user read permission for files.

## Limitations

Functions in `fst` expect a reasonably shallow and small directory structures to deal with, as that is what usually happens in testing. During build-up, tear-down, and comparisons they create collections of filesystem information objects in memory. It is not necessarily the most efficient way, but it allows for more graceful permissions handling.

If you are concerned that you will hold a few copies of full file information lists during the execution, then this library may be a poor match to your needs.

## History: breaking backward compatibility

### v.1 &rarr; v.2

#### No-error signatures

In v.1 many user-facing `fst` funcs returned errors to the caller. While following the general coding practice this pattern did not consider the testing nature of the `fst` package. Most `fst` funcs should abort the test when encountering errors because (a) continuing testing on the wrondly set up data is senseless or (b) cleanup errors indicate failures not caught by the test or a bad setup to begin with. Failing early spares the user from error handling and shortens the trace output for error review.

To accommodate this different philosophy, previously error-returning `fst` funcs are stripped of those return values and a first parameter `fst.Fatalfable` interface added. The likes of `testing.T` and `log.Logger` types satisfy `fst.Fatalfable`. The expected pattern is to pass `t` into `fst` funcs which will call `t.Fatalf` with a reasonable context in a case of an error.

For example:

<table style="text-align: left;">
<tr style="vertical-align: top;"><th>&nbsp;</th><th>v.1.x</th><th>v.2.x</th><tr>
<tr style="vertical-align: top;"><th>signature</th>
<td>

```go
func FileDelAll(root, name string) error
```

</td><td>

```go
func FileDelAll(f Fatalfable, root, name string)
```

</td></tr>
<tr style="vertical-align: top;"><th>usage</th>
<td>

```go
err = fst.FileDelAll("mock", ".gitkeep")
if err != nil {
  t.Fatal(err)
}
```

</td><td>

```go
fst.FileDelAll(t, "mock", ".gitkeep")
```

</td></tr></table>

#### Deprecated `io.Reader` input

Version 1.x of the `fst` package had relied solely on the `io.Reader` interface to feed `fst.TreeCreate` and its derivative funcs with file informaion. While it seemed as a good idea at the time, in practice it provided little utility. With that it had adversely hidden the parsing logic in the `fst.TreeCreate` function.

Versions 2.x breaks the compatibility: instead of `io.Reader` those funcs now expect `[]*Node` - thus allowing better checks at compile time and better testing of the `fst` package.

The provided func `fst.ParseReader` simplifies transition from Version 1.x. It has the parsing logic extracted from `fst.TreeCreate`. Here is an example of using it:

```go
tree := `
2017-11-12T13:14:15Z	0750	settings/
2017-11-12T13:14:15Z	0640	settings/theme1.toml	key = val1
2017-11-12T13:14:15Z	0640	settings/theme2.toml	key = val2
`
nodes := fst.ParseReader(t, strings.NewReader(tree))

old, cleanup := fst.TempCreateChdir(t, nodes)
defer cleanup()
```

Also, see an example of using it while reading the file system nodes' information from a file at [the func `TestTreeDiffTimes`](tree_diff_test.go).

<hr />
