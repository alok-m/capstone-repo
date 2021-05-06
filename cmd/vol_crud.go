package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/alok-m/capstone-codebase/storage"

	"github.com/spf13/cobra"
)

/*
READ
WRITE
DELETE
*/

func init() {
	rootCmd.AddCommand(volCrud)
}

var volCrud = &cobra.Command{
	Use:   "volume",
	Short: "crud for volume",
	Long:  `Volume crud (creating and deleteing)`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "create" {
			//WILL TRUNCATE CARE IF EXISTS
			err := os.Mkdir(args[1], 0755)
			if err != nil {
				log.Fatal(err)
			}
			_, err = storage.NewVolume(args[1], 0)
			if err != nil {
				log.Fatal(err)
			}
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("New volume created at " + args[1] + "\n")
		}
		if args[0] == "delete" {
			err := os.RemoveAll(args[1])
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}
