/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/golangast/lunarr-goast/internal/db/movies"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		//server.Server()

		m, err := movies.GetFiles("assets/video")
		if err != nil {
			panic(err)
		}
		for i := 0; i < len(m); i++ {
			fmt.Println(m[i])
		}

	},
}

func init() {
	rootCmd.AddCommand(startCmd)

}
