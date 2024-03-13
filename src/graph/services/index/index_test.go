package index

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pathutil "rat/graph/util/path"
)

func TestIndex_Add(t *testing.T) {
	t.Parallel()

	type fields struct {
		paths []string
	}

	type args struct {
		path pathutil.NodePath
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantPaths []string
	}{
		{
			"+insertAtEnd",
			fields{[]string{"path1"}},
			args{"zzzzz"},
			[]string{"path1", "zzzzz"},
		},
		{
			"+insertAtStart",
			fields{[]string{"path1"}},
			args{"aaa"},
			[]string{"aaa", "path1"},
		},
		{
			"+insertInMiddle",
			fields{[]string{"path1", "path3"}},
			args{"path2"},
			[]string{"path1", "path2", "path3"},
		},
		{
			"-duplicate",
			fields{[]string{"path1", "path3"}},
			args{"path3"},
			[]string{"path1", "path3"},
		},
		{
			"-empty",
			fields{[]string{}},
			args{"path3"},
			[]string{"path3"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			idx := &Index{
				paths: tt.fields.paths,
			}

			idx.Add(tt.args.path)

			if diff := cmp.Diff(tt.wantPaths, idx.paths); diff != "" {
				t.Errorf("Add() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIndex_Remove(t *testing.T) {
	t.Parallel()

	type fields struct {
		paths []string
	}

	type args struct {
		path pathutil.NodePath
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantPaths []string
	}{
		{
			"+removeAtEnd",
			fields{[]string{"path1", "zzzzz"}},
			args{"zzzzz"},
			[]string{"path1"},
		},
		{
			"+removeAtStart",
			fields{[]string{"aaa", "path1", "zzzzz"}},
			args{"aaa"},
			[]string{"path1", "zzzzz"},
		},
		{
			"+removeInMiddle",
			fields{[]string{"aaa", "path1", "zzzzz"}},
			args{"path1"},
			[]string{"aaa", "zzzzz"},
		},
		{
			"-notFound",
			fields{[]string{"path1", "path3"}},
			args{"ffffff"},
			[]string{"path1", "path3"},
		},
		{
			"-empty",
			fields{[]string{}},
			args{"path3"},
			[]string{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			idx := &Index{
				paths: tt.fields.paths,
			}

			idx.Remove(tt.args.path)

			if diff := cmp.Diff(tt.wantPaths, idx.paths); diff != "" {
				t.Errorf("Remove() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
