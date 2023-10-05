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
	Weight  string `yaml:"weight,omitempty"`
	Content string `yaml:"content,omitempty"`
}

// TemplateData describes data that can be used in node templates.
type TemplateData struct {
	Name string
	Year int
	Week int
}

// NewTemplateData populates template data fields.
func NewTemplateData(name string) *TemplateData {
	year, week := time.Now().ISOWeek()

	return &TemplateData{
		Name: name,
		Year: year,
		Week: week,
	}
}

// FillWeight fills templated node weight.
func (nt *NodeTemplate) FillWeight(td *TemplateData) (int, error) {
	if nt.Weight == "" {
		return 0, nil
	}

	w, err := strconv.Atoi(nt.Weight)
	if err != nil {
		weightTemplate, err := template.New("").Parse(nt.Weight)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse weight template")
		}

		weightBuf := &bytes.Buffer{}

		err = weightTemplate.Execute(weightBuf, td)
		if err != nil {
			return 0, errors.Wrap(err, "failed to execute weight template")
		}

		w, err = strconv.Atoi(weightBuf.String())
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
