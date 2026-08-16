package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func cleanHTML(val string) string {
	val = strings.ReplaceAll(val, "<br>", " ")
	val = strings.ReplaceAll(val, "<br/>", " ")
	val = strings.ReplaceAll(val, "</br>", " ")

	var builder strings.Builder
	builder.Grow(len(val))

	inTag := false
	for i := 0; i < len(val); i++ {
		char := val[i]
		if char == '<' {
			inTag = true
		} else if char == '>' && inTag {
			inTag = false
		} else if !inTag {
			builder.WriteByte(char)
		}
	}

	return strings.Join(strings.Fields(builder.String()), " ")
}

func convertVanshDate(dateStr string) string {
	if dateStr == "" {
		return "N/A"
	}
	parsed, err := time.Parse("Jan 02 2006", dateStr+" 2026")
	if err != nil {
		return dateStr
	}

	days := int(time.Since(parsed).Hours() / 24)
	if days < 0 {
		days = 0
	}
	if days >= 30 {
		return fmt.Sprintf("%dmo", days/30)
	}
	return fmt.Sprintf("%dd", days)
}

func convertZShahDate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" || dateStr == "—" {
		return "N/A"
	}

	parsed, err := time.Parse("Jan 02, 2006", dateStr)
	if err != nil {
		return dateStr
	}

	days := int(time.Since(parsed).Hours() / 24)
	if days < 0 {
		days = 0
	}
	if days >= 30 {
		return fmt.Sprintf("%dmo", days/30)
	}
	return fmt.Sprintf("%dd", days)
}

func convertISODate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" || dateStr == "-" {
		return "N/A"
	}

	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}

	days := int(time.Since(parsed).Hours() / 24)
	if days < 0 {
		days = 0
	}
	if days >= 30 {
		return fmt.Sprintf("%dmo", days/30)
	}
	return fmt.Sprintf("%dd", days)
}

// ageToDays converts relative age strings ("0d", "14d", "1mo") into approximate integer days for chronological sorting.
// Other formats become 9999 to push them to the bottom of the table.
func ageToDays(ageStr string) int {
	clean := strings.TrimSpace(ageStr)

	if strings.HasSuffix(clean, "d") {
		numStr := strings.TrimSuffix(clean, "d")
		if val, err := strconv.Atoi(numStr); err == nil {
			return val
		}
	} else if strings.HasSuffix(clean, "mo") {
		numStr := strings.TrimSuffix(clean, "mo")
		if val, err := strconv.Atoi(numStr); err == nil {
			return val * 30 // Approximate 1 month = 30 days
		}
	}

	return 9999 // Push unknown or "N/A" formats to the very bottom of the table
}

func cleanCompanyName(name string) string {
	name = strings.TrimSpace(name)

	name = strings.ReplaceAll(name, "**", "")
	name = strings.ReplaceAll(name, "🔥 ", "")
	name = strings.ReplaceAll(name, "🔥", "")
	name = strings.ReplaceAll(name, "✓", "")
	name = strings.ReplaceAll(name, "🆁", "")
	name = strings.ReplaceAll(name, "🇺🇸", "")
	name = strings.ReplaceAll(name, "🛂", "")
	name = strings.ReplaceAll(name, "🆕", "")

	return strings.TrimSpace(name)
}
