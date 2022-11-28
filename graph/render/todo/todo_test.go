package todo

import (
	"testing"
	"time"
)

func TestTodoList_Priority(t *testing.T) {
	start := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	// day := 24 * time.Hour

	timeNow = func() time.Time {
		return start
	}

	type fields struct {
		Due  time.Time
		Size time.Duration
	}

	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		// {
		// 	"7 days for 3h task",
		// 	fields{
		// 		Due:  start.Add(7 * day),
		// 		Size: 3 * time.Hour,
		// 	},
		// 	0,
		// },
		// {
		// 	"7 days for 6h task",
		// 	fields{
		// 		Due:  start.Add(7 * day),
		// 		Size: 6 * time.Hour,
		// 	},
		// 	0,
		// },
		// {
		// 	"7 days for 12h task",
		// 	fields{
		// 		Due:  start.Add(7 * day),
		// 		Size: 12 * time.Hour,
		// 	},
		// 	0,
		// },
		// {
		// 	"14 days for 3h task",
		// 	fields{
		// 		Due:  start.Add(14 * day),
		// 		Size: 3 * time.Hour,
		// 	},
		// 	0,
		// },
		// {
		// 	"14 days for 6h task",
		// 	fields{
		// 		Due:  start.Add(14 * day),
		// 		Size: 6 * time.Hour,
		// 	},
		// 	0,
		// },
		// {
		// 	"14 days for 12h task",
		// 	fields{
		// 		Due:  start.Add(14 * day),
		// 		Size: 12 * time.Hour,
		// 	},
		// 	0,
		// },
		// {
		// 	"1 days for 2 day task",
		// 	fields{
		// 		Due:  start.Add(1 * day),
		// 		Size: 2 * day,
		// 	},
		// 	0,
		// },
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tl := &TodoList{
				Due:  tt.fields.Due,
				Size: tt.fields.Size,
			}

			if got := tl.Priority(); got != tt.want {
				t.Errorf("TodoList.Priority() = %v, want %v", got, tt.want)
			}
		})
	}
}
