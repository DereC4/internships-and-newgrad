package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type JobListing struct {
	Company  string
	Role     string
	Location string
	Link     string
	Age      string
}

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

		cols := strings.Split(row, "<td>")
		if len(cols) < 6 {
			continue
		}

		company := cleanHTML(cols[1])
		role := cleanHTML(cols[2])
		location := cleanHTML(cols[3])
		appCell := cols[4]
		age := cleanHTML(cols[5])

		if company == "↳" || company == "" {
			company = lastCompany
		} else {
			lastCompany = company
		}

		var appURL string
		if strings.Contains(appCell, "href=\"") {
			_, afterHref, _ := strings.Cut(appCell, "href=\"")
			appURL, _, _ = strings.Cut(afterHref, "\"")
		}

		if appURL == "" || strings.Contains(appCell, "🔒") {
			continue
		}

		jobs = append(jobs, JobListing{
			Company:  company,
			Role:     role,
			Location: location,
			Link:     appURL,
			Age:      age,
		})
	}

	return jobs
}

func cleanHTML(val string) string {
	val = strings.ReplaceAll(val, "</td>", "")
	val = strings.ReplaceAll(val, "</tr>", "")
	val = strings.ReplaceAll(val, "<strong>", "")
	val = strings.ReplaceAll(val, "</strong>", "")
	return strings.TrimSpace(val)
}

func dogWorker(url string, ch chan string) {
	// function signature in Go is variableName dataType
	// channels are type safe in Go so you have to define what type a channel takes
	response, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprintf("Error fetching %s: %v", url, err)
		return
	}

	bodyBytes, err := io.ReadAll(response.Body)

	if err != nil {
		ch <- fmt.Sprintf("Error reading body for %s: %v", url, err)
		return
	}

	ch <- url + "|derexXD certified separator|" + string(bodyBytes)

}

func main() {
	urls := []string{
		// this is a slice
		"https://raw.githubusercontent.com/vanshb03/Summer2027-Internships/dev/README.md",
		"https://raw.githubusercontent.com/SimplifyJobs/Summer2026-Internships/refs/heads/dev/README.md",
	}

	resultsChannel := make(chan string)
	// make a channel type so we can talk to main
	// loop through urls and start a thread for each one
	fmt.Println("Starting fetches")
	for _, url := range urls {
		// underscore is so we ignore the index, discard it
		go dogWorker(url, resultsChannel)
	}

	// you have to open the file before reading from channel
	file, err := os.OpenFile("testing.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	for i := 0; i < len(urls); i++ {
		rawPayload := <-resultsChannel
		// separate the url from results
		fetchedURL, results, _ := strings.Cut(rawPayload, "|derexXD certified separator|")
		// channels will get consumed when you read them all one by one, so our two for loop approach was writing nothing
		fmt.Printf("--- Document Received #%d from %s ---\n", i+1, fetchedURL)
		fmt.Println(results)
		fmt.Println("-------------------------------")

		separator := fmt.Sprintf("\n\n# --- Document Received #%d ---\n\n", i+1)

		if _, err := file.WriteString(separator); err != nil {
			fmt.Printf("Error writing separator to file: %v\n", err)
		}

		if _, err := file.WriteString(results); err != nil {
			fmt.Printf("Error writing content to file: %v\n", err)
		}

		fmt.Printf("Saved document #%d to combined_output.md\n", i+1)

		if strings.Contains(fetchedURL, "SimplifyJobs") {
			fmt.Println("Processing Simplify Repo...")

			_, afterHeader, foundHeader := strings.Cut(results, "## 💻 Software Engineering Internship Roles")
			if foundHeader {
				sweTable, _, _ := strings.Cut(afterHeader, "</table>")
				sweTable = sweTable + "</table>"

				parsedJobs := parseSimplify(sweTable)
				fmt.Printf("Parsed %d jobs from Simplify!\n", len(parsedJobs))
			}
		} else if strings.Contains(fetchedURL, "vanshb03") {
			fmt.Println("Processing Vansh Repo...")
		}
	}

	for i := 0; i < len(urls); i++ {
	}

}
