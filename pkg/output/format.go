package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// Enhanced color functions
var (
	// Header colors
	headerColor   = color.New(color.FgCyan, color.Bold)
	sectionColor  = color.New(color.FgBlue, color.Bold)
	
	// Status colors
	successColor  = color.New(color.FgGreen, color.Bold)
	errorColor    = color.New(color.FgRed, color.Bold)
	warningColor  = color.New(color.FgYellow, color.Bold)
	infoColor     = color.New(color.FgBlue, color.Bold)
	
	// Detail colors
	dimColor      = color.New(color.Faint)
	emphasisColor = color.New(color.FgMagenta, color.Bold)
	
	// Icons
	successIcon = "✓"
	errorIcon   = "✗"
	warningIcon = "⚠"
	infoIcon    = "ℹ"
	buildIcon   = "🔨"
	packageIcon = "📦"
	timeIcon    = "⏱"
)

// Formatter provides methods for formatted console output
type Formatter struct {
	startTime time.Time
	spinner   *spinner.Spinner
}

// New creates a new formatter
func New() *Formatter {
	return &Formatter{
		startTime: time.Now(),
	}
}

// Header prints a prominent header (Tier 1)
func (f *Formatter) Header(text string) {
	fmt.Printf("\n")
	headerColor.Printf("═══ %s ═══\n", text)
	fmt.Printf("\n")
}

// Section prints a section header (Tier 2)
func (f *Formatter) Section(text string) {
	fmt.Printf("\n")
	sectionColor.Printf("▶ %s\n", text)
}

// Success prints a success message
func (f *Formatter) Success(text string) {
	successColor.Printf("%s %s\n", successIcon, text)
}

// Info prints an informational message
func (f *Formatter) Info(text string) {
	infoColor.Printf("%s %s\n", infoIcon, text)
}

// Warning prints a warning message
func (f *Formatter) Warning(text string) {
	warningColor.Printf("%s %s\n", warningIcon, text)
}

// Error prints an error message
func (f *Formatter) Error(text string) {
	errorColor.Printf("%s %s\n", errorIcon, text)
}

// Detail prints detailed information (indented, dimmed)
func (f *Formatter) Detail(text string) {
	dimColor.Printf("  │ %s\n", text)
}

// Step prints a numbered step
func (f *Formatter) Step(number int, text string) {
	emphasisColor.Printf("[%d] %s\n", number, text)
}

// BuildStart indicates the start of a build operation
func (f *Formatter) BuildStart(packageName string) {
	fmt.Printf("  %s Building %s", buildIcon, packageName)
	f.StartSpinner()
}

// BuildSuccess indicates successful completion of a build
func (f *Formatter) BuildSuccess(packageName string, duration time.Duration) {
	f.StopSpinner()
	successColor.Printf(" %s Built %s", successIcon, packageName)
	dimColor.Printf(" (%v)\n", duration)
}

// BuildError indicates build failure
func (f *Formatter) BuildError(packageName string, err error) {
	f.StopSpinner()
	errorColor.Printf(" %s Failed to build %s: %v\n", errorIcon, packageName, err)
}

// StartSpinner starts a loading spinner
func (f *Formatter) StartSpinner() {
	if f.spinner != nil {
		f.spinner.Stop()
	}
	f.spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	f.spinner.Start()
}

// StopSpinner stops the current spinner
func (f *Formatter) StopSpinner() {
	if f.spinner != nil {
		f.spinner.Stop()
		f.spinner = nil
	}
}

// Progress creates and returns a progress bar
func (f *Formatter) Progress(max int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionEnableColorCodes(true),
	)
}

// Table prints a well-formatted table
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
	sectionColor.Print("  ")
	for i, h := range headers {
		sectionColor.Printf("%-*s  ", widths[i], h)
	}
	fmt.Println()

	// Print separator
	dimColor.Print("  ")
	for i, w := range widths {
		dimColor.Print(strings.Repeat("─", w))
		if i < len(widths)-1 {
			dimColor.Print("  ")
		}
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("  ")
		for i, cell := range row {
			if i < len(widths) {
				fmt.Printf("%-*s  ", widths[i], cell)
			}
		}
		fmt.Println()
	}
}

// Summary prints a build summary with timing information
func (f *Formatter) Summary(successCount, totalCount int, duration time.Duration) {
	fmt.Printf("\n")
	sectionColor.Printf("▶ Build Summary\n")
	
	if successCount == totalCount {
		successColor.Printf("%s All %d packages built successfully ", successIcon, totalCount)
	} else {
		if successCount > 0 {
			successColor.Printf("%s %d packages built successfully ", successIcon, successCount)
		}
		if totalCount - successCount > 0 {
			errorColor.Printf("%s %d packages failed ", errorIcon, totalCount - successCount)
		}
	}
	
	dimColor.Printf("(%s %v)\n", timeIcon, duration)
}

// PackageInfo displays package information in a formatted way
func (f *Formatter) PackageInfo(name, version, packageType string, dependencies []string) {
	fmt.Printf("\n")
	emphasisColor.Printf("%s %s", packageIcon, name)
	if version != "" {
		dimColor.Printf(" v%s", version)
	}
	if packageType != "" {
		dimColor.Printf(" (%s)", packageType)
	}
	fmt.Printf("\n")
	
	if len(dependencies) > 0 {
		f.Detail(fmt.Sprintf("Dependencies: %s", strings.Join(dependencies, ", ")))
	}
}

// DependencyTree prints a simple dependency tree
func (f *Formatter) DependencyTree(packages map[string][]string) {
	f.Section("Dependency Tree")
	
	for pkg, deps := range packages {
		emphasisColor.Printf("├─ %s\n", pkg)
		for i, dep := range deps {
			if i == len(deps)-1 {
				dimColor.Printf("   └─ %s\n", dep)
			} else {
				dimColor.Printf("   ├─ %s\n", dep)
			}
		}
	}
}

// Elapsed returns the time elapsed since the formatter was created
func (f *Formatter) Elapsed() time.Duration {
	return time.Since(f.startTime)
}

// Separator prints a visual separator
func (f *Formatter) Separator() {
	dimColor.Printf("────────────────────────────────────────────────────────────────\n")
}

// NewLine prints a new line
func (f *Formatter) NewLine() {
	fmt.Printf("\n")
}
