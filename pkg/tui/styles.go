// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import "fmt"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

const (
	iconConnected    = "[+]"
	iconDisconnected = "[-]"
	iconWarning      = "[!]"
	iconUnknown      = "[?]"
)

func statusIcon(active, hasError bool) string {
	switch {
	case hasError:
		return iconWarning
	case active:
		return iconConnected
	default:
		return iconDisconnected
	}
}

// FormatStatus returns a colorized status string with an activity icon.
func FormatStatus(active bool, status string) string {
	icon := statusIcon(active, false)
	switch {
	case active && status != "error":
		return colorGreen + icon + " " + status + colorReset
	case !active:
		return colorGray + icon + " " + status + colorReset
	default:
		return colorRed + icon + " " + status + colorReset
	}
}

// FormatHealth returns a colorized health status with score.
func FormatHealth(status string, score int) string {
	colored := func(s string, color string) string {
		return fmt.Sprintf("%s%s (%d/100)%s", color, s, score, colorReset)
	}
	switch status {
	case "excellent":
		return colored(status, colorGreen)
	case "good":
		return colored(status, colorCyan)
	case "fair":
		return colored(status, colorYellow)
	case "poor":
		return colored(status, colorRed)
	default:
		return colorGray + status + colorReset
	}
}

// TruncateString shortens s to maxLen runes, appending "..." when truncated.
func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// PadRight pads s with spaces on the right to the given width, or truncates if longer.
func PadRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	padding := make([]byte, width-len(runes))
	for i := range padding {
		padding[i] = ' '
	}
	return s + string(padding)
}
