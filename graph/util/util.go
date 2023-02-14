package util

import (
	"fmt"
	"net/url"
)

// ReverseSlice reverses a slice.
func ReverseSlice[T any](a []T) []T {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 { //nolint:lll
		a[left], a[right] = a[right], a[left]
	}

	return a
}

// Link returns a markdown link to a node with given path.
func Link(path, name string) string {
	var (
		u url.URL
		q = make(url.Values)
	)

	u.Path = "/view/"

	q.Add("node", path)

	u.RawQuery = q.Encode()

	return fmt.Sprintf("[%s](%s)", name, u.String())
}
