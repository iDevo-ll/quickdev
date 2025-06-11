package utils

// ANSI color codes
const (
	Reset      = "\033[0m"
	Bold       = "\033[1m"
	DimStyle   = "\033[2m"
	Italic     = "\033[3m"
	Underline  = "\033[4m"
	
	// Regular colors
	Black     = "\033[30m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	
	// Bright colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
)

// Style functions for consistent formatting
func Info(s string) string {
	return Blue + s + Reset
}

func Success(s string) string {
	return Green + s + Reset
}

func Warning(s string) string {
	return Yellow + s + Reset
}

func Error(s string) string {
	return Red + s + Reset
}

func Highlight(s string) string {
	return Bold + Cyan + s + Reset
}

func Dimmed(s string) string {
	return BrightBlack + s + Reset
}

func Header(s string) string {
	return Bold + BrightBlue + s + Reset
}

func Section(s string) string {
	return Magenta + s + Reset
}

func Command(s string) string {
	return BrightYellow + s + Reset
}

func Path(s string) string {
	return BrightCyan + s + Reset
}

func Status(s string) string {
	return BrightGreen + s + Reset
} 