package graph

import (
	"bytes"
	"math/rand"
	"strconv"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

//nolint:gochecknoglobals,gosmopolitan // strings contain weird runes.
var smiles = []string{
	"ʕ •́؈•̀)",
	"(╭ರᴥ•́)",
	"┬─┬ ノʕ•ᴥ•ノʔ",
	"₍ᐢ•ﻌ•ᐢ₎",
	"(◕‿◕✿)",
	"(*・‿・)ノ⌒*:･ﾟ✧",
	"(∩ ͡° ͜ʖ ͡°)⊃━☆ﾟ. *",
	"(´・ω・)っ由",
	"~(˘▾˘~)",
	"╰( ⁰ ਊ ⁰ )━☆ﾟ.*･｡ﾟ",
	"＼(-_- )",
	"( ^-^)_旦",
	"(❍ᴥ❍ʋ)",
	"ヽ(͡◕ ͜ʖ ͡◕)ﾉ",
	"｡◕‿‿◕｡",
	"───==≡≡ΣΣ((( つºل͜º)つ",
	"｡◕‿◕｡",
}

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
	RawName string // name that user provided.
	Day     int    // day of the month
	Month   int
	Year    int

	Week    int // week of the year
	YearDay int // day of the year

	Smile string
}

// NewTemplateData populates template data fields.
func NewTemplateData(name string) *TemplateData {
	now := time.Now()

	year, week := now.ISOWeek()
	month := int(now.Month())
	day := now.Day()
	yearDay := now.YearDay()
	smile := smiles[rand.Intn(len(smiles))] //nolint:gosec // week rand num gen.

	return &TemplateData{
		RawTemplateData: RawTemplateData{
			RawName: name,
			Year:    year,
			Month:   month,
			Day:     day,
			Week:    week,
			YearDay: yearDay,
			Smile:   smile,
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
