package token

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_newToken(t *testing.T) {
	t.Parallel()

	type args struct {
		raw string
	}

	tests := []struct {
		args    args
		want    *Token
		wantErr bool
	}{
		{
			args: args{"rat todo"},
			want: &Token{
				Type: TodoTokenType,
				Args: map[string]string{},
			},
			wantErr: false,
		},
		{
			args: args{"rat todo expresion=a=5=c"},
			want: &Token{
				Type: TodoTokenType,
				Args: map[string]string{
					"expresion": "a=5=c",
				},
			},
			wantErr: false,
		},
	}
	for idx, tt := range tests {
		tt := tt

		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			t.Parallel()

			got, err := newToken(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("newToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("newToken() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
