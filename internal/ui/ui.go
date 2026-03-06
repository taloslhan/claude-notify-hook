package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
)

func Info(msg string)    { fmt.Printf("%s%s%s\n", Cyan, msg, Reset) }
func Success(msg string) { fmt.Printf("%s✓ %s%s\n", Green, msg, Reset) }
func Warn(msg string)    { fmt.Printf("%s⚠ %s%s\n", Yellow, msg, Reset) }
func Error(msg string)   { fmt.Printf("%s✗ %s%s\n", Red, msg, Reset) }
func Header(msg string)  { fmt.Printf("\n%s%s%s%s\n", Bold, Blue, msg, Reset) }

func StatusLine(label, value, color string) {
	fmt.Printf("  %-20s %s%s%s\n", label, color, value, Reset)
}

// Prompt asks for input with a default value.
func Prompt(question, defaultVal string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultVal != "" {
		fmt.Printf("%s%s%s [%s]: ", Bold, question, Reset, defaultVal)
	} else {
		fmt.Printf("%s%s%s: ", Bold, question, Reset)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

// Confirm asks a yes/no question.
func Confirm(question string, defaultYes bool) bool {
	hint := "Y/n"
	if !defaultYes {
		hint = "y/N"
	}
	answer := Prompt(fmt.Sprintf("%s (%s)", question, hint), "")
	if answer == "" {
		return defaultYes
	}
	return strings.HasPrefix(strings.ToLower(answer), "y")
}

// MaskToken shows first 4 and last 4 chars of a token.
func MaskToken(token string) string {
	if len(token) <= 10 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
