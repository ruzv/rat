package path

import (
	"net/url"
	"strings"
)

// NodePath filesystem like path that describes where a node is located in the
// graph.
type NodePath string

// NameFromPath returns name of node from its path.
func NameFromPath(path NodePath) string {
	parts := strings.Split(string(path), "/")

	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return parts[len(parts)-1]
}

// PathDepth returns depth.
func PathDepth(path NodePath) int {
	return len(PathParts(path))
}

// PathParts returns path parts.
func PathParts(path NodePath) []string {
	split := strings.Split(string(path), "/")
	parts := make([]string, 0, len(split))

	for _, part := range split {
		if part == "" {
			continue
		}

		parts = append(parts, part)
	}

	return parts
}

// ParentPath returns parent path of node. Returns root path for root path.
func ParentPath(path NodePath) NodePath {
	parts := PathParts(path)

	if len(parts) < 2 {
		return path
	}

	return NodePath(strings.Join(parts[:len(parts)-1], "/"))
}

// JoinName adds the supplied name to the end of the supplied path.
func JoinName(parent NodePath, name string) NodePath {
	return NodePath(strings.Join([]string{string(parent), name}, "/"))
}

// JoinPath adds first and second paths together.
func JoinPath(first, second NodePath) NodePath {
	return NodePath(strings.Join([]string{string(first), string(second)}, "/"))
}

// URL returns a URL to a node with given path.
func URL(path NodePath) string {
	var (
		u url.URL
		q = make(url.Values)
	)

	u.Path = "/view/"

	q.Add("node", string(path))

	u.RawQuery = q.Encode()

	return u.String()
}
