package todo

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	t.Parallel()

	type args struct {
		raw string
	}

	tests := []struct {
		name    string
		args    args
		want    *Todo
		wantErr bool
	}{
		{
			"+valid",
			args{
				strings.Join(
					[]string{
						"- task 1",
						"x task 2",
						"x task 3",
					},
					"\n",
				),
			},
			&Todo{
				Entries: []*Entry{
					{Text: "task 1"},
					{Text: "task 2", Done: true},
					{Text: "task 3", Done: true},
				},
			},
			false,
		},
		{
			"+withHints",
			args{
				strings.Join(
					[]string{
						"priority=1",
						"- task 1",
					},
					"\n",
				),
			},
			&Todo{
				Entries: []*Entry{
					{Text: "task 1"},
				},
				Hints: []*Hint{
					{Type: Priority, Value: 1},
				},
			},
			false,
		},
		{
			"+empty",
			args{""},
			&Todo{},
			false,
		},
		{
			"+hintBetweenTasks",
			args{
				strings.Join(
					[]string{
						"tags=tag1,tag2",
						"- task 1",
						"priority=1",
						"- task 2",
					},
					"\n",
				),
			},
			&Todo{
				Entries: []*Entry{
					{Text: "task 1"},
					{Text: "task 2"},
				},
				Hints: []*Hint{
					{Type: Tags, Value: "tag1,tag2"},
					{Type: Priority, Value: 1},
				},
			},
			false,
		},

		{
			"+multilineTask",
			args{
				strings.Join(
					[]string{
						"- task 1",
						"  description",
						"",
						"",
						"",
						"  continues",
					},
					"\n",
				),
			},
			&Todo{
				Entries: []*Entry{
					{
						Text: strings.Join(
							[]string{
								"task 1",
								"description",
								"",
								"",
								"",
								"continues",
							},
							"\n",
						),
					},
				},
			},
			false,
		},
		{
			"+indentation",
			args{
				strings.Join(
					[]string{
						"- task 1",
						"    description",
						"      description",
						"    description",
						"  description",
					},
					"\n",
				),
			},
			&Todo{
				Entries: []*Entry{
					{
						Text: strings.Join(
							[]string{
								"task 1",
								"  description",
								"    description",
								"  description",
								"description",
							},
							"\n",
						),
					},
				},
			},
			false,
		},
		{
			"-whiteSpace",
			args{
				strings.Join(
					[]string{
						"      ",
						"       ",
						"",
					},
					"\n",
				),
			},
			&Todo{},
			false,
		},
		{
			"-incorrectIndentation",
			args{
				strings.Join(
					[]string{
						"- task 1",
						"description",
					},
					"\n",
				),
			},
			nil,
			true,
		},
		{
			"-invalidHint",
			args{
				strings.Join(
					[]string{
						"invalid=hint",
						"- task 1",
					},
					"\n",
				),
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
