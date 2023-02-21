package util

import (
	"fmt"

	pathutil "private/rat/graph/util/path"
)

// ReverseSlice reverses a slice.
func ReverseSlice[T any](a []T) []T {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 { //nolint:lll
		a[left], a[right] = a[right], a[left]
	}

	return a
}

// Link returns a markdown link to a node with given path.
func Link(path pathutil.NodePath, name string) string {
	return fmt.Sprintf("[%s](%s)", name, pathutil.URL(path))
}
