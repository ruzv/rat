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

// Map applies a function to all todo entries.
func Map[T any, R any](s []T, f func(T) R) []R {
	r := make([]R, len(s))
	for i, v := range s {
		r[i] = f(v)
	}

	return r
}

// Values returns all values of a map.
func Values[K comparable, V any](m map[K]V) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}

	return r
}

// Filter creates a new slice of entries of s that valid function return's
// true to.
func Filter[T any](s []T, valid func(T) bool) []T {
	var r []T

	for _, v := range s {
		if valid(v) {
			r = append(r, v)
		}
	}

	return r
}

// Iter iterates over a slice and calls the given function for each element.
func Iter[T any](s []T, f func(T)) {
	for _, v := range s {
		f(v)
	}
}
