package controllers

import "strings"

func normalizePriority(priority string) string {
	switch strings.TrimSpace(strings.ToLower(priority)) {
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}
