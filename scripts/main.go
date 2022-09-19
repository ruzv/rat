package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	i := 0
	for {
		fmt.Println(i)
		i++

		dir, err := os.ReadDir("./content/notes")
		if err != nil {
			panic(err)
		}

		e := []string{
			".metadata.json",
			"content.md",
			"docs",
			"life",
			"loadero",
			"projects",
		}

		if len(dir) != len(e) {
			err = errors.New(fmt.Sprintf("len mistake: %d", len(dir)))
		}

		for idx, d := range dir {
			fmt.Println(d.Name())
			if d.Name() != e[idx] {
				err = errors.New(
					fmt.Sprintf("name missmatch %s: %s", d.Name(), e[idx]),
				)
			}
		}

		if err != nil {
			panic(err)
		}

	}
}
