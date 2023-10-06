package graph

import (
	"bytes"
	"strconv"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

// NodeTemplate describes a template data of a node.
type NodeTemplate struct {
	Name    string `yaml:"name,omitempty"`
	Weight  string `yaml:"weight,omitempty"`
	Content string `yaml:"content,omitempty"`
}

// TemplateData describes data that can be used in node templates.
type TemplateData struct {
	RawTemplateData
	// name that was filled by name template.
	Name string
}

// RawTemplateData describes the template fields available to unprocessed
// template fillers, like FillName.
type RawTemplateData struct {
	// name that user provided.
	RawName string
	Year    int
	Week    int
}

// NewTemplateData populates template data fields.
func NewTemplateData(name string) *TemplateData {
	year, week := time.Now().ISOWeek()

	return &TemplateData{
		RawTemplateData: RawTemplateData{
			RawName: name,
			Year:    year,
			Week:    week,
		},
		// name is only populated after name template is executed.
	}
}

// FillName fills templated node name.
func (nt *NodeTemplate) FillName(td *RawTemplateData) (string, error) {
	nameTemplate, err := template.New("").Parse(nt.Name)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse name template")
	}

	buff := &bytes.Buffer{}

	err = nameTemplate.Execute(buff, td)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute name template")
	}

	return buff.String(), nil
}

// FillWeight fills templated node weight.
func (nt *NodeTemplate) FillWeight(td *TemplateData) (int, error) {
	w, err := strconv.Atoi(nt.Weight)
	if err != nil {
		weightTemplate, err := template.New("").Parse(nt.Weight)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse weight template")
		}

		buff := &bytes.Buffer{}

		err = weightTemplate.Execute(buff, td)
		if err != nil {
			return 0, errors.Wrap(err, "failed to execute weight template")
		}

		w, err = strconv.Atoi(buff.String())
		if err != nil {
			return 0, errors.Wrap(err, "failed to convert weight to int")
		}
	}

	return w, nil
}

// FillContent fills templated node content.
func (nt *NodeTemplate) FillContent(td *TemplateData) (string, error) {
	buff := &bytes.Buffer{}

	contentTemplate, err := template.New("").Parse(nt.Content)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse content template")
	}

	err = contentTemplate.Execute(buff, td)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute template")
	}

	return buff.String(), nil
}
