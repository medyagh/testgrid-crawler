package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ProwJob represents a single job run from the Prow history
type ProwJob struct {
	SpyglassLink string `json:"SpyglassLink"`
	ID           string `json:"ID"`
	Started      string `json:"Started"`
	Duration     int64  `json:"Duration"` // Nanoseconds
	Result       string `json:"Result"`
}

// FetchJobHistory scrapes the Prow job history page for a given job name
// It follows pagination to fetch more jobs.
func FetchJobHistory(jobName string) ([]ProwJob, error) {
	baseURL := fmt.Sprintf("https://prow.k8s.io/job-history/gs/kubernetes-ci-logs/logs/%s", jobName)
	nextURL := baseURL
	var allJobs []ProwJob
	
	// Safety limit to prevent infinite loops (e.g., 20 pages or ~400 jobs)
	maxPages := 20 

	for i := 0; i < maxPages; i++ {
		jobs, nextLink, err := fetchPage(nextURL)
		if err != nil {
			return nil, err
		}
		
		allJobs = append(allJobs, jobs...)

		if nextLink == "" {
			break
		}
		
		// Construct next URL
		// nextLink is relative, usually likeStub: "?buildId=..." or "/job-history/..."
		// The HTML shows: /job-history/gs/kubernetes-ci-logs/logs/ci-minikube-integration?buildId=...
		// But let's handle just the query part if possible, or full path.
		
		// If it's a full path (starting with /), prepend domain
		if strings.HasPrefix(nextLink, "/") {
			nextURL = "https://prow.k8s.io" + nextLink
		} else if strings.HasPrefix(nextLink, "?") {
			nextURL = baseURL + nextLink
		} else {
			// Fallback or errors
			break
		}
	}

	return allJobs, nil
}

func fetchPage(pageURL string) ([]ProwJob, string, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch URL %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("received status code %d from %s", resp.StatusCode, pageURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read body: %w", err)
	}

	// 1. Extract JSON
	re := regexp.MustCompile(`var allBuilds = (\[.*?\]);`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, "", fmt.Errorf("could not find allBuilds JSON in response")
	}

	var jobs []ProwJob
	if err := json.Unmarshal(matches[1], &jobs); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// 2. Extract Next Link (Older Runs)
	// <a href="/job-history/gs/kubernetes-ci-logs/logs/ci-minikube-integration?buildId=2008545318748033024"><- Older Runs</a>
	// We look for "Older Runs" text and the href associated with it.
	// A simple regex might be sufficient.
	// match href="(...)"...Older Runs
	nextLinkRe := regexp.MustCompile(`<a href="([^"]+)"[^>]*>&lt;- Older Runs</a>`)
	linkMatches := nextLinkRe.FindSubmatch(body)
	var nextLink string
	if len(linkMatches) >= 2 {
		nextLink = string(linkMatches[1])
		// Decode HTML entities if needed, but usually simple URL is fine.
		// However, the anchor text might be different or have classes.
		// The example was: <td><a href="...">...</a></td>
	}
	
	return jobs, nextLink, nil
}

// FormatDuration converts nanoseconds to a human-readable string
func FormatDuration(ns int64) string {
	d := time.Duration(ns)
	return d.Round(time.Second).String()
}
