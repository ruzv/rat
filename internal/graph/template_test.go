package graph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNodeTemplate_FillNames(t *testing.T) {
	t.Parallel()

	type fields struct {
		DisplayName string
		PathName    string
		Weight      string
		Content     string
		Template    *NodeTemplate
	}

	type args struct {
		td *RawTemplateData
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		wantName     string
		wantFilename string
		wantErr      bool
	}{
		{
			"+valid",
			fields{
				DisplayName: "{{ .RawName }}",
			},
			args{&RawTemplateData{
				RawName: "test",
			}},
			"test",
			"test",
			false,
		},
		{
			"+withFilenameTemplate",
			fields{
				DisplayName: "{{ .RawName }}",
				PathName:    "{{ .RawName }} filename",
			},
			args{&RawTemplateData{
				RawName: "test",
			}},
			"test",
			"test-filename",
			false,
		},
		{
			"+fixedName",
			fields{
				DisplayName: "🌊 Current (week {{ .Week }})",
				PathName:    "ashtasht",
			},
			args{&RawTemplateData{
				Week: 4,
			}},
			"🌊 Current (week 4)",
			"ashtasht",
			false,
		},
		{
			"+onlyDisplayNameFromTemplateField",
			fields{
				DisplayName: "{{ .Day }}",
			},
			args{&RawTemplateData{
				Day: 4,
			}},
			"4",
			"4",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nt := &NodeTemplate{
				DisplayName: tt.fields.DisplayName,
				PathName:    tt.fields.PathName,
				Weight:      tt.fields.Weight,
				Content:     tt.fields.Content,
				Template:    tt.fields.Template,
			}

			name, filename, err := nt.FillNames(tt.args.td)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"NodeTemplate.FillNames() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if diff := cmp.Diff(tt.wantName, name); diff != "" {
				t.Fatalf(
					"NodeTemplate.FillNames() got missmatch (-want +got):\n%s",
					diff,
				)
			}

			if diff := cmp.Diff(tt.wantFilename, filename); diff != "" {
				t.Fatalf(
					"NodeTemplate.FillNames() got1 missmatch (-want +got):\n%s",
					diff,
				)
			}
		})
	}
}
