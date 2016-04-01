// dupes finds duplicate files in the given root directory
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// hashes maps from digests to paths
var hashes = make(map[string]string)

// hasher is used by checksum to calculate digests
var hasher = sha256.New()

// files counts the number of files examined
var files int64

// dupes counts the number of duplicate files
var dupes int64

// wasted counts the space (in bytes) occupied by duplicates
var wasted int64

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

	if !info.Mode().IsRegular() || info.Size() == 0 {
		return nil
	}

	files++

	sum, err := checksum(path)
	if err != nil {
		return err
	}

	if dupe, ok := hashes[sum]; ok {
		fmt.Printf("%s\n%s\n\n", path, dupe)
		dupes++
		wasted += info.Size()
	} else {
		hashes[sum] = path
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("error: need an argument")
		os.Exit(1)
	}

	filepath.Walk(os.Args[1], check)

	if len(hashes) > 0 {
		fmt.Printf("%d files examined, %d duplicates found, %d bytes wasted\n", files, dupes, wasted)
	}
}
