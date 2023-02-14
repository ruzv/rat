package path

import "strings"

// NameFromPath returns name of node from its path.
func NameFromPath(path string) string {
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return parts[len(parts)-1]
}

// PathDepth returns depth.
func PathDepth(path string) int {
	return len(PathParts(path))
}

// PathParts returns path parts.
func PathParts(path string) []string {
	split := strings.Split(path, "/")
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
func ParentPath(path string) string {
	parts := PathParts(path)

	if len(parts) < 2 {
		return path
	}

	return strings.Join(parts[:len(parts)-1], "/")
}
