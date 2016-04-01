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
*regular* files. It'll print pairs of paths, separated by an empty line,
for each duplicate it finds across all roots. It'll also print statistics
at the end:

```
$ dupes ~/Downloads/Stuff/

...

/home/phf/Downloads/Stuff/Maps/107031a.jpg
/home/phf/Downloads/Stuff/107031.jpg

/home/phf/Downloads/Stuff/Maps/bwpoliticalmap2.gif
/home/phf/Downloads/Stuff/Maps/bwpoliticalmap.gif

/home/phf/Downloads/Stuff/Maps/calendar_awesome.jpg
/home/phf/Downloads/Stuff/Maps/393031_3485829032040_1460844946_3050753_389238284_n.jpg

/home/phf/Downloads/Stuff/Maps/cogh-undercity.gif
/home/phf/Downloads/Stuff/Denis/cogh-undercity.gif

...

2,301 files examined, 87 duplicates found, 126.14 MB wasted
```

There's even a `-p` option for using "paranoid" byte-by-byte file comparisons
instead of SHA256 digests. (As a bonus it'll warn you about any SHA256
collisions it finds in "paranoid" mode. You should feel very lucky indeed
if you actually get one of those.)

## License

The MIT License.

## Random Notes

- I tried different hash functions (MD5, SHA1, SHA256, SHA512) but none had
the upper hand in terms of performance; I ended up removing the code to make
them configurable to keep the tool simple; the default, SHA256, is probably
overkill in terms of reliability, but what the heck; it's plenty fast.
