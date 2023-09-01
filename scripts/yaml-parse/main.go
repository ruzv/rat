package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func main() {
	var fileName string

	cmd := &cobra.Command{
		Use:   "yaml-parse",
		Short: "check parsing markdown files with yaml header",
		Run: func(_ *cobra.Command, _ []string) {
			content, err := os.ReadFile(fileName)
			if err != nil {
				panic(err)
			}

			type values map[string]interface{}

			header := &struct {
				Any      values    `yaml:",inline"`
				ID       uuid.UUID `yaml:"id"`
				Template string    `yaml:"template"`
			}{}

			err = yaml.Unmarshal(content, header)
			if err != nil {
				panic(err)
			}

			b, err := json.MarshalIndent(header, "", "  ")
			if err != nil {
				panic(err)
			}

			fmt.Println(string(b))
		},
	}

	cmd.Flags().StringVarP(&fileName, "file", "f", "", "file path to parse")
	err := cmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}

	err = cmd.Execute()
	if err != nil {
		panic(err)
	}
}
