package formatter

import "github.com/fatih/color"

func colorizeStatus(status string) string {
	switch status {
	case "pending":
		return color.YellowString(status)
	case "in_progress":
		return color.CyanString(status)
	case "completed":
		return color.GreenString(status)
	case "blocked":
		return color.RedString(status)
	default:
		return status
	}
}
