# dupes: find duplicate files

A quick hack to find duplicate files. Since I am trying to practice Go again,
guess what, it's written in Go. But don't get overly excited, okay?

## Install

You should be able to just say this:

	go install github.com/phf/dupes

If you have `$GOPATH/bin` in your `$PATH` you can then just run `dupes`
directly.

## Usage

It's very simple:

	dupes path-to-root

The command will walk the directories under `path-to-root` and examine all
*regular* files. It'll print pairs of paths, separated by an empty line,
for each duplicate it finds. It'll also print some statistics at the end:

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

2301 files examined, 87 duplicates found, 132268638 bytes wasted
```

Nothing to it.

## License

The MIT License.
