package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		password string
		cost     int
	)
	cli := &cobra.Command{
		Use:   "gen-paswd-hash",
		Short: "generate a password hash",
		Long:  "generate a password hash to be used for graph owners auth",
		Run: func(cmd *cobra.Command, args []string) {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
			if err != nil {
				panic(err)
			}

			fmt.Println(hash)
			fmt.Println(string(hash))
		},
	}

	cli.Flags().
		StringVarP(&password, "password", "p", "", "password to hash")
	cli.MarkFlagRequired("password")

	cli.Flags().
		IntVarP(&cost, "cost", "c", bcrypt.MinCost, "cost of bcrypt hash function")

	cli.Execute()
}
