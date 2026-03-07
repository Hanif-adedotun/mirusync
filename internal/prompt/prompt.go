package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func String(prompt string, defaultVal string) (string, error) {
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
