package main

import "strings"

func parseSimplify(rawHTML string) []JobListing {
	var jobs []JobListing
	var lastCompany string

	_, afterTbody, foundStart := strings.Cut(rawHTML, "<tbody>")
	tbodyContent, _, foundEnd := strings.Cut(afterTbody, "</tbody>")
	if !foundStart || !foundEnd {
		return nil
	}

	rows := strings.Split(tbodyContent, "<tr>")

	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		row = strings.ReplaceAll(row, "<td align=\"center\">", "<td>")

		cols := strings.Split(row, "<td>")
		if len(cols) < 6 {
			continue
		}

		appCell := cols[4]

		if strings.Contains(appCell, "🔒") {
			continue
		}

		var appURL string
		if strings.Contains(appCell, "href=\"") {
			_, afterHref, _ := strings.Cut(appCell, "href=\"")
			appURL, _, _ = strings.Cut(afterHref, "\"")
		}

		if appURL == "" {
			continue
		}

		company := cleanHTML(cols[1])
		role := cleanHTML(cols[2])
		location := cleanHTML(cols[3])
		age := cleanHTML(cols[5])

		if company == "↳" || company == "" {
			company = lastCompany
		} else {
			lastCompany = company
		}

		jobs = append(jobs, JobListing{
			Company:  company,
			Role:     role,
			Location: location,
			Link:     appURL,
			Age:      age,
			NewGrad:  false,
		})
	}

	return jobs
}

func parseVansh(rawMarkdown string) []JobListing {
	var jobs []JobListing
	var lastCompany string

	// discard the before return string, we do not need that
	_, afterStart, foundStart := strings.Cut(rawMarkdown, "<!-- Please leave a one line gap between this and the table TABLE_START (DO NOT CHANGE THIS LINE) -->")
	tableContent, _, foundEnd := strings.Cut(afterStart, "[⬆️ Back to Top ⬆️]")

	if !foundStart || !foundEnd {
		return nil
	}

	lines := strings.Split(tableContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") {
			continue
		}

		cols := strings.Split(line, "|")
		if len(cols) < 6 {
			continue
		}

		// cols[0] = "" (prefix)
		// cols[1] = Company
		// cols[2] = Role
		// cols[3] = Location
		// cols[4] = Application / Link
		// cols[5] = Date Posted
		appCell := cols[4]

		if strings.Contains(appCell, "🔒") {
			continue
		}

		var appURL string
		if strings.Contains(appCell, "href=\"") {
			_, afterHref, _ := strings.Cut(appCell, "href=\"")
			appURL, _, _ = strings.Cut(afterHref, "\"")
		}

		if appURL == "" {
			continue
		}

		company := cleanHTML(cols[1])
		role := cleanHTML(cols[2])
		location := cleanHTML(cols[3])
		datePosted := convertVanshDate(cleanHTML(cols[5]))

		if company == "↳" || company == "" {
			company = lastCompany
		} else {
			lastCompany = company
		}

		jobs = append(jobs, JobListing{
			Company:  company,
			Role:     role,
			Location: location,
			Link:     appURL,
			Age:      datePosted,
			NewGrad:  false,
		})
	}

	return jobs
}

func parseSpeedyApply(rawMarkdown string) []JobListing {
	var jobs []JobListing

	sections := []struct{ start, end string }{
		{"<!-- TABLE_FAANG_START -->", "<!-- TABLE_FAANG_END -->"},
		{"<!-- TABLE_QUANT_START -->", "<!-- TABLE_QUANT_END -->"},
		{"<!-- TABLE_START -->", "<!-- TABLE_END -->"},
	}

	for _, sec := range sections {
		_, afterStart, foundStart := strings.Cut(rawMarkdown, sec.start)
		if !foundStart {
			continue
		}

		tableContent, _, foundEnd := strings.Cut(afterStart, sec.end)
		if !foundEnd {
			continue
		}

		for _, line := range strings.Split(tableContent, "\n") {
			line = strings.TrimSpace(line)

			if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") || strings.Contains(line, "Company") {
				continue
			}

			cols := strings.Split(line, "|")
			if len(cols) < 7 {
				continue
			}

			var appCell, ageCell string

			if len(cols) >= 8 {
				appCell = cols[5]
				ageCell = cols[6]
			} else {
				// cols[1]=Company, cols[2]=Role, cols[3]=Location, cols[4]=Link, cols[5]=Age
				appCell = cols[4]
				ageCell = cols[5]
			}

			if strings.Contains(appCell, "🔒") {
				continue
			}

			_, afterHref, foundHref := strings.Cut(appCell, "href=\"")
			if !foundHref {
				continue
			}
			appURL, _, _ := strings.Cut(afterHref, "\"")

			jobs = append(jobs, JobListing{
				Company:  cleanHTML(cols[1]),
				Role:     cleanHTML(cols[2]),
				Location: cleanHTML(cols[3]),
				Link:     appURL,
				Age:      cleanHTML(ageCell),
				NewGrad:  false,
			})
		}
	}

	return jobs
}
