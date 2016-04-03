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
	"strings"
)

// bytesize ints represent a size as in "bytes of memory"
type bytesize uint64

// String formats the underlying integer with suitable units
// (KB, MB, .., YB) to keep the number itself small-ish.
func (bs bytesize) String() string {
	units := []string{"bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	value := float64(bs)
	unit := 0
	for value > 1024.0 && unit < len(units)-1 {
		value /= 1024.0
		unit++
	}
	return fmt.Sprintf("%.2f %s", value, units[unit])
}

// countin ints represent a count
type countin uint64

// String formats the underlying integer with commas as "thousands separators"
// to make it easier to read.
func (ci countin) String() string {
	str := fmt.Sprintf("%d", ci)
	chunks := splitFromBack(str, 3)
	str = strings.Join(chunks, ",")
	return str
}

// split string s into chunks of at most n characters from the back.
func splitFromBack(s string, n int) []string {
	var chunks []string

	fullChunks := len(s) / n
	restChunk := len(s) % n

	if restChunk > 0 {
		chunks = append(chunks, s[0:restChunk])
	}

	var i = restChunk
	for fullChunks > 0 {
		chunks = append(chunks, s[i:i+n])
		i += n
		fullChunks--
	}
	return chunks
}

var paranoid = flag.Bool("p", false, "paranoid byte-by-byte comparison")
var goroutines = flag.Int("j", runtime.NumCPU(), "number of goroutines")

// hashes maps from digests to paths
var hashes = make(map[string]string)

// sizes maps from sizes to paths
var sizes = make(map[int64]string)

// hasher is used by checksum to calculate digests
var hasher = sha256.New()

// files counts the number of files examined
var files countin

// dupes counts the number of duplicate files
var dupes countin

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
	wasted += bytesize(size)

	return nil
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

	for _, root := range flag.Args() {
		filepath.Walk(root, check)
	}

	if len(sizes) > 0 || len(hashes) > 0 {
		fmt.Printf("%v files examined, %v duplicates found, %v wasted\n", files, dupes, wasted)
	}
}
