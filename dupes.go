// dupes finds duplicate files in the given root directory
//
// TODO concurrent checksum/compare? library for other programs?
package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

var paranoid = flag.Bool("p", false, "paranoid byte-by-byte comparison")
var goroutines = flag.Int("j", runtime.NumCPU(), "number of goroutines")

// hashes maps from digests to paths
var hashes = make(map[string]string)

// sizes maps from sizes to paths
var sizes = make(map[int64]string)

// hasher is used by checksum to calculate digests
var hasher = sha256.New()

// files counts the number of files examined
var files int64

// dupes counts the number of duplicate files
var dupes int64

// wasted counts the space (in bytes) occupied by duplicates
var wasted int64

// identical does a byte-by-byte comparison of the files with the
// given paths
func identical(pa, pb string) (bool, error) {
	const bufferSize = 4096

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

	// TODO: move these out to avoid reallocating them? lazy make?
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

	hasher.Reset()
	io.Copy(hasher, file)
	sum := fmt.Sprintf("%x", hasher.Sum(nil))

	return sum, nil
}

// check is called for each path we walk. It only examines regular, non-empty
// files. For each file it calculates a checksum; if it has seen the same
// checksum before, it signals a duplicate; otherwise it remembers the
// checksum and the path of the original file before moving on.
func check(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	size := info.Size()

	if !info.Mode().IsRegular() || size == 0 {
		return nil
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

	fmt.Printf("%s\n%s\n\n", path, dupe)
	dupes++
	wasted += size

	return nil
}

func main() {
	flag.Usage = func() {
		var program = os.Args[0]
		fmt.Fprintf(os.Stderr, "Usage: %s [option]... directory...\n", program)
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
	}

	for _, root := range flag.Args() {
		filepath.Walk(root, check)
	}

	if len(sizes) > 0 || len(hashes) > 0 {
		fmt.Printf("%d files examined, %d duplicates found, %d bytes wasted\n", files, dupes, wasted)
	}
}
