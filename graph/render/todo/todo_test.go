package todo

import (
	"testing"
	"time"
)

func TestTodoList_Priority(t *testing.T) {
	type fields struct {
		Due  time.Time
		Size time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			"in future",
			fields{
				Due:  time.Now().Add(61 * 24 * time.Hour),
				Size: 3 * time.Hour * 24,
			},
			0,
		},
		{
			"past due",
			fields{
				Due:  time.Now().Add(-24 * time.Hour),
				Size: time.Hour * 5,
			},
			0,
		},
	}
	for _, tt := range tests {
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
