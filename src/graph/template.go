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
	RawName   string // name that user provided.
	Day       int    // day of the month   1,2,..31
	Month     int    // month of the year 1,2,3...12
	MonthName string // month name in English
	Year      int

	Week        int    // week number of the year
	Weekday     int    // number of the weekday 1,2...7
	WeekdayName string // monday (1), tuesday(2)...sunday
	YearDay     int    // day of the year 1,2,...365

	// next ++ increment of weight of largest sibling node
	WeightAutoincrement int

	Smile string
}

// Template prepares n nodes template. Walking up the graph tree to find
// a base template and setting template data to be used to fill template fields.
func (n *Node) Template(
	p Provider,
	rawName string,
) (*NodeTemplate, *TemplateData, error) {
	nt, err := n.GetTemplate(p)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get template")
	}

	leafs, err := n.GetLeafs(p)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get leafs")
	}

	maxWeight := 0

	if len(leafs) != 0 {
		for _, leaf := range leafs {
			if leaf.Header.Weight > maxWeight {
				maxWeight = leaf.Header.Weight
			}
		}
	}

	now := time.Now()

	year, week := now.ISOWeek()
	weekday := now.Weekday()
	month := now.Month()
	day := now.Day()
	yearDay := now.YearDay()
	smile := smiles[rand.Intn(len(smiles))] //nolint:gosec // week rand num gen.
	weightAutoincrement := maxWeight + 1

	rawTemplateData := RawTemplateData{
		RawName:             rawName,
		Year:                year,
		Month:               int(month),
		MonthName:           month.String(),
		Day:                 day,
		Week:                week,
		Weekday:             (int(weekday) + 7) % 7,
		WeekdayName:         weekday.String(),
		YearDay:             yearDay,
		WeightAutoincrement: weightAutoincrement,
		Smile:               smile,
	}

	displayName, pathName, err := nt.FillNames(&rawTemplateData)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to fill names")
	}

	return nt, &TemplateData{
		RawTemplateData: rawTemplateData,
		DisplayName:     displayName,
		PathName:        pathName,
	}, nil
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
