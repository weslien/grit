package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "grit",
	Short: "Go-based monorepo tool",
	Long:  "GRIT - Go Monorepo Tool with advanced dependency management and build caching",
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
