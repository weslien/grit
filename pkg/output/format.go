package output

import (
	"fmt"
	"strings"
	_ "strings"
)

// Color codes for terminal output
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	// Foreground colors
	Black  = "\033[30m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	// Add this constant if it's not already defined

	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

// Formatter provides methods for formatted console output
type Formatter struct{}

// New creates a new formatter
func New() *Formatter {
	return &Formatter{}
}

// Header prints a prominent header (Tier 1)
func (f *Formatter) Header(text string) {
	fmt.Printf("\n%s%s%s%s\n\n", Bold, Cyan, text, Reset)
}

// Section prints a section header (Tier 2)
func (f *Formatter) Section(text string) {
	fmt.Printf("\n%s%s%s%s\n", Bold, Blue, text, Reset)
}

// Success prints a success message (Tier 2)
func (f *Formatter) Success(text string) {
	fmt.Printf("%s%s✓ %s%s\n", Bold, Green, text, Reset)
}

// Info prints an informational message (Tier 2)
func (f *Formatter) Info(text string) {
	fmt.Printf("%s%sℹ %s%s\n", Bold, Blue, text, Reset)
}

// Warning prints a warning message (Tier 2)
func (f *Formatter) Warning(text string) {
	fmt.Printf("%s%s⚠ %s%s\n", Bold, Yellow, text, Reset)
}

// Error prints an error message (Tier 2)
func (f *Formatter) Error(text string) {
	fmt.Printf("%s%s✗ %s%s\n", Bold, Red, text, Reset)
}

// Detail prints detailed information (Tier 3)
func (f *Formatter) Detail(text string) {
	fmt.Printf("  %s%s%s\n", Dim, text, Reset)
}

// Add these methods to the Formatter type to support the example_output.go file

// Step prints a step in a process
func (f *Formatter) Step(number int, text string) {
	fmt.Printf("%s%s[%d]%s %s\n", Bold, Magenta, number, Reset, text)
}

// Table prints a simple table with headers and rows
func (f *Formatter) Table(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	fmt.Print(Bold)
	for i, h := range headers {
		fmt.Printf("%-*s", widths[i]+2, h)
	}
	fmt.Println(Reset)

	// Print separator
	for _, w := range widths {
		fmt.Print(strings.Repeat("─", w+2))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Printf("%-*s", widths[i]+2, cell)
			}
		}
		fmt.Println()
	}
}
