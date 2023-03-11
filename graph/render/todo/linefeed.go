package todo

type lineFeed struct {
	lines   []string
	current int
}

func (lf *lineFeed) next() bool {
	return lf.current < len(lf.lines)
}

func (lf *lineFeed) pop() string {
	if !lf.next() {
		return ""
	}

	defer func() { lf.current++ }()

	return lf.lines[lf.current]
}

func (lf *lineFeed) peek() string {
	if !lf.next() {
		return ""
	}

	return lf.lines[lf.current]
}

func (lf *lineFeed) popUntil(cond func(string) bool) []string {
	var lines []string

	for lf.next() {
		if cond(lf.peek()) {
			break
		}

		lines = append(lines, lf.pop())
	}

	return lines
}
