package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Prefix and spacing for init-style prompts (chevron left of question, spaced out).
const (
	Chevron = "  > "
	Circle  = "  ○ "
)

func lineBreakAndPrefix(prefix string) {
	fmt.Println()
	fmt.Print(prefix)
}

func String(prompt string, defaultVal string) (string, error) {
	return StringStyled(prompt, defaultVal, "", false)
}

// StringStyled prompts for a string. If prefix is non-empty, prints a blank line and prefix before the question.
func StringStyled(prompt string, defaultVal string, prefix string, spaced bool) (string, error) {
	if spaced && prefix != "" {
		lineBreakAndPrefix(prefix)
	}
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", scanner.Err()
	}
	s := strings.TrimSpace(scanner.Text())
	if s == "" && defaultVal != "" {
		return defaultVal, nil
	}
	return s, nil
}

func Int(prompt string, defaultVal int) (int, error) {
	return IntStyled(prompt, defaultVal, "", false)
}

func IntStyled(prompt string, defaultVal int, prefix string, spaced bool) (int, error) {
	if spaced && prefix != "" {
		lineBreakAndPrefix(prefix)
	}
	fmt.Printf("%s [%d]: ", prompt, defaultVal)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return 0, scanner.Err()
	}
	s := strings.TrimSpace(scanner.Text())
	if s == "" {
		return defaultVal, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func Select(prompt string, options []string, defaultIdx int) (int, error) {
	return SelectStyled(prompt, options, defaultIdx, "", false)
}

func SelectStyled(prompt string, options []string, defaultIdx int, prefix string, spaced bool) (int, error) {
	if spaced && prefix != "" {
		lineBreakAndPrefix(prefix)
	}
	for i, o := range options {
		fmt.Printf("  %d) %s\n", i+1, o)
	}
	def := defaultIdx + 1
	if def < 1 || def > len(options) {
		def = 1
	}
	fmt.Printf("%s [%d]: ", prompt, def)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return 0, scanner.Err()
	}
	s := strings.TrimSpace(scanner.Text())
	if s == "" {
		return defaultIdx, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > len(options) {
		return 0, fmt.Errorf("choose 1–%d", len(options))
	}
	return n - 1, nil
}

func Confirm(prompt string, defaultYes bool) (bool, error) {
	return ConfirmStyled(prompt, defaultYes, "", false)
}

func ConfirmStyled(prompt string, defaultYes bool, prefix string, spaced bool) (bool, error) {
	if spaced && prefix != "" {
		lineBreakAndPrefix(prefix)
	}
	def := "y"
	if !defaultYes {
		def = "n"
	}
	fmt.Printf("%s [%s]: ", prompt, def)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false, scanner.Err()
	}
	s := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if s == "" {
		return defaultYes, nil
	}
	return s == "y" || s == "yes", nil
}

func Pause(msg string) {
	if msg != "" {
		fmt.Println(msg)
	}
	fmt.Print("Press Enter to continue...")
	bufio.NewScanner(os.Stdin).Scan()
}
