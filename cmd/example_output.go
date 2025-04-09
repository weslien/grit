package cmd

import (
	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/output"
)

var exampleOutputCmd = &cobra.Command{
	Use:   "example-output",
	Short: "Example of formatted output",
	Run: func(cmd *cobra.Command, args []string) {
		fmt := output.New()
		
		fmt.Header("GRIT - Modern Monorepo Tool")
		
		fmt.Section("Building Packages")
		fmt.Step(1, "Resolving dependencies")
		fmt.Detail("Found 5 packages with 3 dependencies")
		
		fmt.Step(2, "Compiling packages")
		fmt.Success("Built package: common")
		fmt.Success("Built package: utils")
		fmt.Warning("Package api has outdated dependencies")
		fmt.Success("Built package: api")
		
		fmt.Section("Test Results")
		fmt.Info("Running tests for all packages")
		fmt.Success("All tests passed")
		
		fmt.Section("Summary")
		headers := []string{"Package", "Status", "Time"}
		rows := [][]string{
			{"common", "Success", "0.5s"},
			{"utils", "Success", "0.3s"},
			{"api", "Success", "1.2s"},
		}
		fmt.Table(headers, rows)
	},
}

func init() {
	rootCmd.AddCommand(exampleOutputCmd)
}