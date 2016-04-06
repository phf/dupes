// Command dupes finds duplicate files in the given root directory
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
	paranoid    = flag.Bool("p", false, "paranoid byte-by-byte file comparison")
	minimumSize = flag.Int64("s", 1, "minimum size (in bytes) of files to consider")
	globbing    = flag.String("g", globDefault, "glob expression for files to consider")
)

var (
	hashes = make(map[string]string)   // maps from digests to paths
	sizes  = make(map[int64]string)    // maps from sizes to paths
	final  = make(map[string][]string) // maps from paths to duplicate paths (collates all dupes)

	files  counter  // number of files examined
	dupes  counter  // number of duplicate files
	wasted bytesize // space (in bytes) occupied by duplicates
)

// fileContentsMatch does a byte-by-byte comparison of the files with the
// given paths
func fileContentsMatch(pa, pb string) (bool, error) {
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

	return fileContentsHelper(a, b)
}

func fileContentsHelper(a, b io.Reader) (bool, error) {
	bufferSize := os.Getpagesize()

	ba := make([]byte, bufferSize)
	bb := make([]byte, bufferSize)

	for {
		la, erra := a.Read(ba)
		lb, errb := b.Read(bb)

		// specification of Read() says to check returned size
		// before considering errors; who are we to disagree?
		if la > 0 || lb > 0 {
			// it's okay to use Equal here because whatever may
			// be left behind in the buffers after a short read
			// had to be Equal in the prior iteration; note that
			// we don't have to check la == lb either because if
			// they were not, Equal should fail
			if !bytes.Equal(ba, bb) {
				return false, nil
			}
		}

		// specification of Read() says that sooner or later we'll
		// see io.EOF regardless of returned size; only if both
		// files end in the same iteration (and made it past Equal
		// above) do we have a duplicate
		switch {
		case erra == io.EOF && errb == io.EOF:
			return true, nil
		case erra != nil:
			return false, erra
		case errb != nil:
			return false, errb
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

	if *globbing != globDefault {
		matched, err := filepath.Match(*globbing, info.Name())
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
		same, err := fileContentsMatch(path, dupe)
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
	for k := range final {
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

	_, err := filepath.Match(*globbing, "checking pattern syntax")
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
