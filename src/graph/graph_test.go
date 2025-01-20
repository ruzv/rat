package graph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_parsePathName(t *testing.T) {
	t.Parallel()

	type args struct {
		name string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"+valid",
			args{"name"},
			"name",
		},
		{
			"+numbers",
			args{"name1234"},
			"name1234",
		},
		{
			"+separator",
			args{"name-1234_attt"},
			"name-1234-attt",
		},
		{
			"-dissalowed",
			args{`name:;;'"#$%`},
			"name",
		},
		{
			"-emoji",
			args{` 🎥 Viewed `},
			"viewed",
		},
		{
			"-spaces",
			args{`Dijkstra’s shortest path`},
			"dijkstras-shortest-path",
		},
		{
			"-spacesAndForbiddenChars",
			args{` 🎥 Dijkstra’s  🎥  shortest  🎥  path  🎥  `},
			"dijkstras-shortest-path",
		},
		{
			"-dashes",
			args{`Boy - The Heron`},
			"boy-the-heron",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parsePathName(tt.args.name)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("parsePathName() = %s", diff)
			}
		})
	}
}
