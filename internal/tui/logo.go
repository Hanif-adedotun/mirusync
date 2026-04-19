package tui

import "fmt"

// ANSI color codes (compatible with most terminals)
const (
	ColorReset   = "\033[0m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorDim     = "\033[2m"
)

// ASCII wordmark logo for "MIRUSYNC" with magenta/cyan split.
func PrintLogo() {
	// Custom ASCII glyphs keep "Y" visually distinct from "U" across terminals.
	logo := []string{
		" __  __ ___ ____  _   _ ____  __   __ _   _  ____ ",
		"|  \\/  |_ _|  _ \\| | | / ___| \\ \\ / /| \\ | |/ ___|",
		"| |\\/| || || |_) | | | \\___ \\  \\ V / |  \\| | |    ",
		"| |  | || ||  _ <| |_| |___) |  | |  | |\\  | |___ ",
		"|_|  |_|___|_| \\_\\\\___/|____/   |_|  |_| \\_|\\____|",
	}
	// Color: first half of each line magenta, second half cyan
	for _, line := range logo {
		runes := []rune(line)
		mid := len(runes) / 2
		fmt.Println(ColorMagenta + string(runes[:mid]) + ColorCyan + string(runes[mid:]) + ColorReset)
	}
	fmt.Println()
	fmt.Println(ColorDim + "  Sync folders between machines over SSH" + ColorReset)
	fmt.Println()
}
