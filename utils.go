package main

import (
	"fmt"
	"strings"
)

// formatSizeWithUnit formats the given uint (which represents a size as
// in "number of bytes") with suitable units (KB, MB, .., YB) to keep the
// number itself small-ish.
func formatSizeWithUnit(size uint64) string {
	units := []string{"bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	value := float64(size)

	u := 0
	for value > 1024.0 && u < len(units)-1 {
		value /= 1024.0
		u++
	}

	return fmt.Sprintf("%.2f %s", value, units[u])
}

// formatCountWithThousands formats the given uint with commas as
// "thousands separators" to make it easier to read.
func formatCountWithThousands(count uint64) string {
	str := fmt.Sprintf("%d", count)
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

// just so we can attach a String method
type bytesize uint64

// String formats with formatSizeWithUnit.
func (b bytesize) String() string {
	return formatSizeWithUnit(uint64(b))
}

// just so we can attach a String method
type counter uint64

// String formats with formatCountWithThousands.
func (c counter) String() string {
	return formatCountWithThousands(uint64(c))
}
