package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	// These are the four colors we actually use throughout the tool.
	// Keeping them in one place means changing the look is a one-liner.
	green  = color.New(color.FgGreen, color.Bold)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed, color.Bold)
	cyan   = color.New(color.FgCyan)
	dim    = color.New(color.FgHiBlack)
	white  = color.New(color.FgWhite, color.Bold)
)

// Success prints a ✓ success message.
func Success(format string, a ...any) {
	green.Fprintf(os.Stdout, "  ✓  "+format+"\n", a...)
}

// Info prints a general informational line.
func Info(format string, a ...any) {
	cyan.Fprintf(os.Stdout, "  →  "+format+"\n", a...)
}

// Warn prints a warning that doesn't stop execution.
func Warn(format string, a ...any) {
	yellow.Fprintf(os.Stdout, "  ⚠  "+format+"\n", a...)
}

// Error prints an error message to stderr.
func Error(format string, a ...any) {
	red.Fprintf(os.Stderr, "  ✗  "+format+"\n", a...)
}

// Dim prints subtle/secondary information.
func Dim(format string, a ...any) {
	dim.Fprintf(os.Stdout, "     "+format+"\n", a...)
}

// Header prints a section header with a top border.
func Header(title string) {
	fmt.Println()
	white.Printf("  %s\n", title)
	dim.Printf("  %s\n", strings.Repeat("─", len(title)+2))
}

// CommitRow prints a single commit in the standard log format.
// Example:  a3f1c9b  feat: add user authentication  (2 days ago)  Alice
func CommitRow(hash, subject, author string, ts time.Time) {
	age := formatAge(ts)

	// truncate long subjects so the output stays on one line
	if len(subject) > 62 {
		subject = subject[:59] + "..."
	}

	fmt.Printf(
		"  %s  %s  %s  %s\n",
		dim.Sprintf("%.7s", hash),
		fmt.Sprintf("%-65s", subject),
		dim.Sprintf("%-16s", age),
		dim.Sprintf("%s", author),
	)
}

// DiffMessage shows an old→new message comparison side by side.
func DiffMessage(old, new string) {
	Header("Message diff")

	fmt.Print("  ")
	red.Print("before: ")
	dim.Println(old)

	fmt.Print("  ")
	green.Print("after:  ")
	fmt.Println(new)
}

// Confirm prints a warning and asks the user to type "yes" to proceed.
// Returns false if the user declines or inputs anything other than "yes".
func Confirm(prompt string) bool {
	yellow.Printf("\n  ⚠  %s\n", prompt)
	fmt.Print("     Type 'yes' to continue: ")

	var input string
	fmt.Scanln(&input) //nolint:errcheck
	return strings.TrimSpace(strings.ToLower(input)) == "yes"
}

// Separator prints a blank divider line.
func Separator() {
	fmt.Println()
}

// formatAge turns a time.Time into a human-readable "N days ago" string.
func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%d min ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%d hours ago", h)
	case d < 48*time.Hour:
		return "yesterday"
	case d < 30*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / (24 * 30))
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(d.Hours() / (24 * 365))
		return fmt.Sprintf("%d years ago", years)
	}
}
