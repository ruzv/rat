package main

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"rat/logr"
)

func main() {
	var (
		password string
		cost     int
		log      = logr.NewLogR(
			os.Stdout,
			"rat-gen-paswd-hash",
			logr.LogLevelDebug,
		)
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

			log.Infof("password hash:\n%s", string(hash))
		},
	}

	cli.Flags().
		StringVarP(&password, "password", "p", "", "password to hash")

	err := cli.MarkFlagRequired("password")
	if err != nil {
		panic(err)
	}

	cli.Flags().IntVarP(
		&cost, "cost", "c", bcrypt.MinCost, "cost of bcrypt hash function",
	)

	err = cli.Execute()
	if err != nil {
		panic(err)
	}
}
