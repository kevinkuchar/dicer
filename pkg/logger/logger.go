package logger

import (
	"fmt"
	"strings"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// formatArgs converts variadic arguments to a single string joined with " - "
func formatArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, arg := range args {
		if i > 0 {
			builder.WriteString(" ")
		}

		switch v := arg.(type) {
		case string:
			builder.WriteString(v)
		case int:
			builder.WriteString(fmt.Sprintf("%d", v))
		case int8:
			builder.WriteString(fmt.Sprintf("%d", v))
		case int16:
			builder.WriteString(fmt.Sprintf("%d", v))
		case int32:
			builder.WriteString(fmt.Sprintf("%d", v))
		case int64:
			builder.WriteString(fmt.Sprintf("%d", v))
		case uint:
			builder.WriteString(fmt.Sprintf("%d", v))
		case uint8:
			builder.WriteString(fmt.Sprintf("%d", v))
		case uint16:
			builder.WriteString(fmt.Sprintf("%d", v))
		case uint32:
			builder.WriteString(fmt.Sprintf("%d", v))
		case uint64:
			builder.WriteString(fmt.Sprintf("%d", v))
		case float32:
			builder.WriteString(fmt.Sprintf("%g", v))
		case float64:
			builder.WriteString(fmt.Sprintf("%g", v))
		case bool:
			builder.WriteString(fmt.Sprintf("%t", v))
		default:
			builder.WriteString(fmt.Sprintf("%v", v))
		}
	}

	return builder.String()
}

// Colored log helpers
func LogInfo(args ...interface{}) {
	msg := formatArgs(args...)
	fmt.Printf("%s[INFO]%s %s\n", ColorCyan, ColorReset, msg)
}

func LogSuccess(args ...interface{}) {
	msg := formatArgs(args...)
	fmt.Printf("%s[SUCCESS]%s %s\n", ColorGreen, ColorReset, msg)
}

func LogWarning(args ...interface{}) {
	msg := formatArgs(args...)
	fmt.Printf("%s[WARNING]%s %s\n", ColorYellow, ColorReset, msg)
}

func LogError(args ...interface{}) {
	msg := formatArgs(args...)
	fmt.Printf("%s[ERROR]%s %s\n", ColorRed, ColorReset, msg)
}

func LogBreak() {
	fmt.Printf("%s--------------------------------%s\n", ColorPurple, ColorReset)
}
