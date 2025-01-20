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
	DisplayName string        `yaml:"displayName,omitempty"`
	PathName    string        `yaml:"pathName,omitempty"`
	Weight      string        `yaml:"weight,omitempty"`
	Content     string        `yaml:"content,omitempty"`
	Template    *NodeTemplate `yaml:"template,omitempty"`
}

// TemplateData describes data that can be used in node templates.
type TemplateData struct {
	RawTemplateData
	// fields populated when template is executed.
	DisplayName string
	PathName    string
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

// FillNames returns templated nodes display name and path name.
func (nt *NodeTemplate) FillNames(td *RawTemplateData) (string, string, error) {
	templt, err := template.New("").Parse(nt.DisplayName)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse name template")
	}

	buff := &bytes.Buffer{}

	err = templt.Execute(buff, td)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to execute name template")
	}

	displayName := buff.String()
	pathName := displayName

	if nt.PathName != "" {
		templt, err := template.New("").Parse(nt.PathName)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to parse name template")
		}

		buff := &bytes.Buffer{}

		err = templt.Execute(buff, td)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to execute name template")
		}

		pathName = buff.String()
	}

	pathName = parsePathName(pathName)
	if pathName == "" {
		return "", "", errors.New("empty path name")
	}

	return displayName, pathName, nil
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
