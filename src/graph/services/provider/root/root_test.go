package root

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"rat/graph"
)

func TestConfig_fillDefaults(t *testing.T) {
	t.Parallel()

	type fields struct {
		Content  string
		Template *graph.NodeTemplate
	}

	tests := []struct {
		name   string
		fields fields
		want   Config
	}{
		{
			"+valid",
			fields{
				Content: "content",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Content: "content",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
		},
		{
			"+defaultContent",
			fields{
				Content: "",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Content: defaultConfig.Content,
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
		},
		{
			"+defaultTemplate",
			fields{
				Content: "name",
			},
			Config{
				Content:  "name",
				Template: defaultConfig.Template,
			},
		},
		{
			"+defaultTemplateName",
			fields{
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     "",
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     defaultConfig.Template.Name,
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
		},
		{
			"+defaultTemplateWeight",
			fields{
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   defaultConfig.Template.Weight,
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
		},
		{
			"+defaultTemplateContent",
			fields{
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  "",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  defaultConfig.Template.Content,
					Template: &graph.NodeTemplate{},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &Config{
				Content:  tt.fields.Content,
				Template: tt.fields.Template,
			}
			got := c.fillDefaults()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Config.fillDefaults() = %s", diff)
			}
		})
	}
}
