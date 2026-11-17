package main

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func isNewGradSource(url string) bool {
	normalizedURL := strings.ToLower(url)
	normalizedURL = strings.ReplaceAll(normalizedURL, "-", "")
	normalizedURL = strings.ReplaceAll(normalizedURL, "_", "")
	return strings.Contains(normalizedURL, "newgrad")
}

func compareJobListings(a, b JobListing) int {
	daysA := ageToDays(a.Age)
	daysB := ageToDays(b.Age)

	if ageComparison := cmp.Compare(daysA, daysB); ageComparison != 0 {
		return ageComparison
	}

	return cmp.Compare(a.Company, b.Company)
}

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

func convertUnixDate(timestamp int64) string {
	if timestamp == 0 {
		return "N/A"
	}

	days := int(time.Since(time.Unix(timestamp, 0)).Hours() / 24)
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

func convertZapplyHours(ageStr string) string {
	clean := strings.TrimSpace(ageStr)
	if !strings.HasSuffix(clean, "h") {
		return ageStr
	}

	numStr := strings.TrimSuffix(clean, "h")
	if _, err := strconv.Atoi(numStr); err != nil {
		return ageStr
	}

	return "0d"
}

func convertZapplyMinutes(ageStr string) string {
	clean := strings.TrimSpace(ageStr)
	if strings.HasSuffix(clean, "mo") || !strings.HasSuffix(clean, "m") {
		return ageStr
	}

	numStr := strings.TrimSuffix(clean, "m")
	if _, err := strconv.Atoi(numStr); err != nil {
		return ageStr
	}

	return "0d"
}

func convertZapplyWeeks(ageStr string) string {
	clean := strings.TrimSpace(ageStr)
	if !strings.HasSuffix(clean, "w") {
		return ageStr
	}

	numStr := strings.TrimSuffix(clean, "w")
	val, err := strconv.Atoi(numStr)
	if err != nil {
		return ageStr
	}

	return fmt.Sprintf("%dd", val*7)
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

func classifyRole(role, sourceCategory string) string {
	if sourceCategory != "" {
		return sourceCategory
	}
	r := strings.ToLower(role)
	if strings.Contains(r, "quant") || strings.Contains(r, "trading") || strings.Contains(r, "trader") || strings.Contains(r, "portfolio") || strings.Contains(r, "hft") || strings.Contains(r, "algorithmic") {
		return "Quant"
	}
	if strings.Contains(r, "machine learning") || strings.Contains(r, "deep learning") || strings.Contains(r, "artificial intelligence") || strings.Contains(r, "computer vision") || strings.Contains(r, "applied scientist") || strings.Contains(r, "research scientist") || strings.Contains(r, "ai engineer") || strings.Contains(r, "ml engineer") || strings.Contains(r, "ai intern") || strings.Contains(r, "ml intern") {
		return "AI_ML"
	}
	if strings.Contains(r, "data scien") || strings.Contains(r, "data analyst") || strings.Contains(r, "business intelligence") || strings.Contains(r, "data engineer") || strings.Contains(r, "analytics") || strings.Contains(r, "bi engineer") || strings.Contains(r, "bi developer") {
		return "Data"
	}
	if strings.Contains(r, "software") || strings.Contains(r, "developer") || strings.Contains(r, "swe") || strings.Contains(r, "frontend") || strings.Contains(r, "backend") || strings.Contains(r, "full stack") || strings.Contains(r, "fullstack") || strings.Contains(r, "mobile") || strings.Contains(r, "ios") || strings.Contains(r, "android") || strings.Contains(r, "devops") || strings.Contains(r, "sre") || strings.Contains(r, "cloud") || strings.Contains(r, "security") || strings.Contains(r, "firmware") || strings.Contains(r, "embedded") || strings.Contains(r, "systems") {
		return "SWE"
	}
	return "Other"
}
