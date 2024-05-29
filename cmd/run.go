/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"lazysync/application"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs an application",
	Long:  `Starts configured application in dedicated role`,
	Run: func(cmd *cobra.Command, args []string) {
		app := application.InitFromConfig()
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
