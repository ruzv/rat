package path

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// NodePath filesystem like path that describes where a node is located in the
// graph.
type NodePath string

// String returns string representation of path.
func (p NodePath) String() string {
	return string(p)
}

// Parent returns parent path of node. Returns self when depth = 1.
func (p NodePath) Parent() NodePath {
	parts := p.Parts()

	if len(parts) < 2 {
		return p
	}

	return NodePath(strings.Join(parts[:len(parts)-1], "/"))
}

// Depth returns depth of path.
func (p NodePath) Depth() int {
	return len(p.Parts())
}

// JoinName adds the supplied name to the end of the supplied path.
func (p NodePath) JoinName(name string) NodePath {
	if p == "" {
		return NodePath(name)
	}

	return NodePath(string(p) + "/" + name)
}

// ViewURL returns a URL to a node with given path.
func (p NodePath) ViewURL() (string, error) {
	u, err := url.JoinPath("/view/", string(p))
	if err != nil {
		return "", errors.Wrap(err, "failed to join path")
	}

	return u, nil
}

// Name returns name of node from its path.
func (p NodePath) Name() string {
	parts := p.Parts()

	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return parts[len(parts)-1]
}

// Parts returns path parts.
func (p NodePath) Parts() []string {
	split := strings.Split(string(p), "/")
	parts := make([]string, 0, len(split))

	for _, part := range split {
		if part == "" {
			continue
		}

		parts = append(parts, part)
	}

	return parts
}
