package internshipsandnewgrad

import "fmt"

func main() {
	urls := []string{
		// this is a slice
		"https://raw.githubusercontent.com/vanshb03/Summer2027-Internships/dev/README.md",
	}

	resultsChannel := make(chan string)
	// make a channel type so we can talk to main
	// loop through urls and start a thread for each one
	fmt.Println("Starting fetches")
	for _, url := range urls {
		// underscore is so we ignore the index, discard it
	}

	for i := 0; i < len(urls); i++ {
		results := <-resultsChannel
		fmt.Printf("--- Document Received #%d ---\n", i+1)
		fmt.Println(results)
		fmt.Println("-------------------------------\n")
	}
}
