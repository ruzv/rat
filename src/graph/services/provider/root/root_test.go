package root

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"rat/graph"
)

func TestConfig_fillDefaults(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name     string
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
				Name:    "rootName",
				Content: "content",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Name:    "rootName",
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
			"+defaultName",
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
				Name:    defaultConfig.Name,
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
				Name:    "rootName",
				Content: "",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Name:    "rootName",
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
				Name:    "rootName",
				Content: "name",
			},
			Config{
				Name:     "rootName",
				Content:  "name",
				Template: defaultConfig.Template,
			},
		},
		{
			"+defaultTemplateName",
			fields{
				Name:    "rootName",
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     "",
					Weight:   "weight",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Name:    "rootName",
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
				Name:    "rootName",
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "",
					Content:  "content",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Name:    "rootName",
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
				Name:    "rootName",
				Content: "name",
				Template: &graph.NodeTemplate{
					Name:     "name",
					Weight:   "weight",
					Content:  "",
					Template: &graph.NodeTemplate{},
				},
			},
			Config{
				Name:    "rootName",
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &Config{
				Name:     tt.fields.Name,
				Content:  tt.fields.Content,
				Template: tt.fields.Template,
			}

			got := c.fillDefaults()

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf(
					"Config.fillDefaults() mismatch (-want +got):\n%s",
					diff,
				)
			}
		})
	}
}
