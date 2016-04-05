// dupes finds duplicate files in the given root directory
package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const (
	globDefault = "*"
)

var (
	paranoid    = flag.Bool("p", false, "paranoid byte-by-byte comparison")
	minimumSize = flag.Int64("s", 1, "minimum size (in bytes) of duplicate file")
	glob        = flag.String("g", globDefault, "glob expression for file names")
)

// hashes maps from digests to paths
var hashes = make(map[string]string)

// sizes maps from sizes to paths
var sizes = make(map[int64]string)

// final maps from paths to duplicate paths (collates all dupes)
var final = make(map[string][]string)

// files counts the number of files examined
var files counter

// dupes counts the number of duplicate files
var dupes counter

// wasted counts the space (in bytes) occupied by duplicates
var wasted bytesize

// identical does a byte-by-byte comparison of the files with the
// given paths
func identical(pa, pb string) (bool, error) {
	bufferSize := os.Getpagesize()

	a, err := os.Open(pa)
	if err != nil {
		return false, err
	}
	defer a.Close()
	b, err := os.Open(pb)
	if err != nil {
		return false, err
	}
	defer b.Close()

	ba := make([]byte, bufferSize)
	bb := make([]byte, bufferSize)

	for {
		la, erra := a.Read(ba)
		lb, errb := b.Read(bb)

		if erra != nil || errb != nil {
			if erra == io.EOF && errb == io.EOF {
				return true, nil
			}
			if erra != nil {
				return false, erra
			}
			if errb != nil {
				return false, errb
			}
		}

		if la != lb { // TODO: short read always at end of file?
			return false, nil
		}

		if !bytes.Equal(ba, bb) {
			return false, nil
		}
	}
}

// checksum calculates a hash digest for the file with the given path
func checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	io.Copy(hasher, file)
	sum := fmt.Sprintf("%x", hasher.Sum(nil))

	return sum, nil
}

// check is called for each path we walk. It only examines regular, non-empty
// files. It first rules out duplicates by file size; for files that remain
// it calculates a checksum; if it has seen the same checksum before, it
// signals a duplicate; otherwise it remembers the checksum and the path of
// the original file before moving on; in paranoid mode it follows up with a
// byte-by-byte file comparison.
func check(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	size := info.Size()

	if !info.Mode().IsRegular() || size < *minimumSize {
		return nil
	}

	if *glob != globDefault {
		matched, err := filepath.Match(*glob, info.Name())
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}
	}

	files++

	var dupe string
	var ok bool
	if dupe, ok = sizes[size]; !ok {
		sizes[size] = path
		return nil
	}

	// backpatch new file into hashes
	sum, err := checksum(dupe)
	if err != nil {
		return err
	}
	hashes[sum] = dupe

	sum, err = checksum(path)
	if err != nil {
		return err
	}

	if dupe, ok = hashes[sum]; !ok {
		hashes[sum] = path
		return nil
	}

	if *paranoid {
		same, err := identical(path, dupe)
		if err != nil {
			return err
		}
		if !same {
			fmt.Printf("cool: %s sha256-collides with %s!\n", path, dupe)
			return nil
		}
	}

	dupes++
	wasted += bytesize(size)

	final[dupe] = append(final[dupe], path)

	return nil
}

func sortedDupes() []string {
	var sk []string
	for k, _ := range final {
		sk = append(sk, k)
	}
	sort.Strings(sk)
	return sk
}

func main() {
	flag.Usage = func() {
		var program = os.Args[0]
		fmt.Fprintf(os.Stderr, "Usage: %s [option...] directory...\n", program)
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
	}

	_, err := filepath.Match(*glob, "checking pattern syntax")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid pattern for -g (%v)\n", err)
		os.Exit(1)
	}

	for _, root := range flag.Args() {
		filepath.Walk(root, check)
	}

	sk := sortedDupes()
	for _, k := range sk {
		vs := final[k]
		fmt.Println(k)
		for _, v := range vs {
			fmt.Println(v)
		}
		fmt.Println()
	}

	if len(sizes) > 0 || len(hashes) > 0 {
		fmt.Printf("%v files examined, %v duplicates found, %v wasted\n", files, dupes, wasted)
	}
}
