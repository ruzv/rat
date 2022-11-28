package graph

import "testing"

func TestParentPath(t *testing.T) {
	t.Parallel()

	type expect struct {
		//
	}

	type args struct {
		path string
	}

	tests := []struct {
		name   string
		expect expect
		args   args
		want   string
	}{
		{
			name:   "+valid",
			expect: expect{},
			args:   args{"root"},
			want:   "root",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ParentPath(tt.args.path)
			if got != tt.want {
				t.Fatalf("ParentPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
