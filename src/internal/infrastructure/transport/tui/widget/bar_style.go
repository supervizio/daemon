// Package widget provides reusable TUI components.
package widget

// BarStyle defines the characters for a progress bar.
// It specifies the fill, empty, and bracket characters used in rendering.
type BarStyle struct {
	Full  string
	Empty string
	Left  string
	Right string
}

// Bar style presets for different visual representations.
var (
	// BlockBar uses block characters.
	BlockBar BarStyle = BarStyle{
		Full:  "█",
		Empty: " ",
		Left:  "",
		Right: "",
	}

	// BracketBar uses brackets with block fill.
	BracketBar BarStyle = BarStyle{
		Full:  "█",
		Empty: " ",
		Left:  "[",
		Right: "]",
	}

	// ASCIIBar uses ASCII characters only.
	ASCIIBar BarStyle = BarStyle{
		Full:  "#",
		Empty: "-",
		Left:  "[",
		Right: "]",
	}

	// SubBlockChars provides 1/8th granularity for progress bars.
	// Index 0 = empty, 1-7 = partial fills, 8 = full.
	SubBlockChars []string = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}

	// Sparks are the sparkline characters (8 levels).
	Sparks []string = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
)
