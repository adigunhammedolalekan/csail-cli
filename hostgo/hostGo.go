package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)
var rootCmd = &cobra.Command{
	Use:   "hostgo",
	Short: "Cloud hosting for Go web applications",
}
func main() {
	appsCmd()
	authCmd()
	envCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
