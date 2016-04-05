# dupes: find duplicate files

[![Go Report Card](https://goreportcard.com/badge/github.com/phf/dupes)](https://goreportcard.com/report/github.com/phf/dupes)

A quick hack to find duplicate files. Since I am trying to practice Go again,
guess what, it's written in Go. But don't get overly excited, okay?

## Install

You should be able to just say this:

	go install github.com/phf/dupes

If you have `$GOPATH/bin` in your `$PATH` you can then just run `dupes`
directly.

## Usage

It's very simple:

	dupes rootpath1 rootpath2 ...

The command will walk the directories under each rootpath and examine all
*regular* files. It'll print clusters of paths, separated by an empty line,
for each duplicate it finds across all roots. It'll also print statistics
at the end:

```
$ dupes ~/Downloads/

...

/home/phf/Downloads/BPT.pdf
/home/phf/Downloads/MATH_BOOKS/Basic_Probability_Theory_Robert_Ash.pdf

/home/phf/Downloads/CharSheets-G(1).pdf
/home/phf/Downloads/CharSheets-G.pdf
/home/phf/Downloads/Stuff/dan/CharSheets-G.pdf

/home/phf/Downloads/fish-0003/README.dist
/home/phf/Downloads/fish-0014/README.dist

...

2,301 files examined, 87 duplicates found, 126.14 MB wasted
```

There's a `-p` option for using "paranoid" byte-by-byte file comparisons
instead of SHA256 digests. (As a bonus it'll warn you about any SHA256
collisions it finds in "paranoid" mode. You should feel very lucky indeed
if you actually get one of those.)

The `-s` option can be used to set a minimum file size you care about; it
defaults to `1` so empty files are ignored.

The `-g` option allows you to specify a
[globbing](https://golang.org/pkg/path/filepath/#Match) pattern for the
file names you care about; it defaults to `*` which matches all file names;
note that you may have to escape the pattern as in `-g '*.pdf'` if the
current directory contains files that would match (which would cause your
shell to do the expansion instead).

## License

The MIT License.

## TODO

- make the darn thing concurrent so we can hide latencies and take advantage
of multiple cores
- wrap it up as a library or service for other Go programs?
- add hard linking or deleting? probably not
- display size of dupes? sort output by size?

## Random Notes

- I tried different hash functions (MD5, SHA1, SHA256, SHA512) but none had
the upper hand in terms of performance; I ended up removing the code to make
them configurable to keep the tool simple; the default, SHA256, is probably
overkill in terms of reliability, but what the heck; it's plenty fast.
- I have to unlearn "sequential performance instincts" like "allocate once
globally" because they don't apply if the things you're allocating now get
written to from multiple goroutines; see `hasher` in `checksum` and the two
buffers in `identical` for instance.
- There are many ways to make `dupes` more concurrent. The obvious one is to
start a goroutine for each root directory. Curiously there's not much of a
speedup if we only pass two or three roots; once we pass twenty or so,
however, things start to heat up.
- Another approach splits `check` into a pipeline of goroutines connected
with channels. Each goroutine "filters out" things that cannot possibly
be duplicates and passes the rest further down to more thorough filters.
(This can be combined with one goroutine for each root directory.
And we can have pools of workers to do long file operations concurrently.)
- I'll have to write a number of competing concurrent variants to see what's
best. So I'll leave `dupes.go` as a non-concurrent reference version for now.

## Kudos

There are *so many* similar programs around, some of them even in Go.
I didn't quite realize just how many there were when I started, yikes.
Here I'll just list the ones that actually inspired a feature. (Maybe
later I'll add some benchmarks as well.)

- https://github.com/mathieuancelin/duplicates minimum file size and regular
expressions for file names (although I decided to use
[globbing](https://golang.org/pkg/path/filepath/#Match) instead)
