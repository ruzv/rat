package graph

import (
	pathutil "rat/graph/util/path"
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
			"name-1234_attt",
		},
		{
			"-dissalowed",
			args{`name:;;'"#$%`},
			"name",
		},
		{
			"-emoji",
			args{` 🎥 Viewed `},
			"Viewed",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parsePathName(tt.args.name)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("parsePathName() = %s", diff)
			}
		})
	}
}

func TestNode_Name(t *testing.T) {
	t.Parallel()

	type fields struct {
		Path    pathutil.NodePath
		Header  NodeHeader
		Content string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			n := &Node{
				Path:    tt.fields.Path,
				Header:  tt.fields.Header,
				Content: tt.fields.Content,
			}
			got := n.Name()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Node.Name() = %s", diff)
			}
		})
	}
}
