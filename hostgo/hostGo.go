package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hostgo",
		Short: "Cloud hosting for Go web applications",
	}
	rootCmd.AddCommand(authCmd(), appsCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
