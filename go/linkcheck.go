package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var errZapplyClosed = errors.New("zapply listing closed")

var linkClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 40,
		MaxConnsPerHost:     40,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if isClosedZapplyBoard(req.URL) {
			return errZapplyClosed
		}

		if !isZapplyHost(req.URL.Host) {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

func isZapplyHost(host string) bool {
	return strings.Contains(strings.ToLower(host), "zapply.jobs")
}

func isClosedZapplyBoard(u *url.URL) bool {
	if !isZapplyHost(u.Host) {
		return false
	}
	path := strings.Trim(strings.ToLower(u.Path), "/")
	return path == "jobs"
}

func checkLink(urlStr string) bool {

	if !isZapplyHost(urlStr) {
		return true
	}

	resp, err := linkClient.Get(urlStr)

	if errors.Is(err, errZapplyClosed) {
		fmt.Printf("removed: %s \n", urlStr)
		return false
	}

	// prevent accidental delete idk we'll do this later
	if err != nil {
		return true
	}

	// go will only reuse connection if response body was read to the eof
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if isClosedZapplyBoard(resp.Request.URL) {
		fmt.Printf("removed: %s -> %s\n", urlStr, resp.Request.URL.String())
		return false
	}

	return true
}

func keepLiveJobs(jobs []JobListing) []JobListing {
	alive := make([]JobListing, 0, len(jobs))
	var zapply []JobListing
	for _, job := range jobs {
		if !isZapplyHost(job.Link) {
			alive = append(alive, job)
			continue
		}
		zapply = append(zapply, job)
	}

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	workers := make(chan struct{}, 40)

	for _, job := range zapply {
		wg.Add(1)
		go func(job JobListing) {
			defer wg.Done()

			workers <- struct{}{}
			keep := checkLink(job.Link)
			<-workers

			if !keep {
				return
			}

			mu.Lock()
			alive = append(alive, job)
			mu.Unlock()
		}(job)
	}

	wg.Wait()
	return alive
}
