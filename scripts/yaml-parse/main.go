package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func main() {
	var fileName string

	cmd := &cobra.Command{
		Use:   "yaml-parse",
		Short: "check parsing markdown files with yaml header",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(fileName)
		},
	}

	cmd.PersistentFlags().
		StringVarP(&fileName, "file", "f", "", "file path to parse")
	cmd.MarkFlagRequired("file")

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
