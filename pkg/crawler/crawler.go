package crawler

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

// Config holds the configuration for the crawler
type Config struct {
	JobName  string
	MaxPages int
}

// Crawler is responsible for scraping Prow job history
type Crawler struct {
	config Config
}

// New creates a new Crawler with the given configuration
func New(config Config) *Crawler {
	if config.MaxPages <= 0 {
		config.MaxPages = 20 // Default limit
	}
	return &Crawler{config: config}
}

// Run scrapes the Prow job history based on the crawler's configuration
func (c *Crawler) Run() ([]ProwJob, error) {
	baseURL := fmt.Sprintf("https://prow.k8s.io/job-history/gs/kubernetes-ci-logs/logs/%s", c.config.JobName)
	nextURL := baseURL
	var allJobs []ProwJob

	for i := 0; i < c.config.MaxPages; i++ {
		jobs, nextLink, err := fetchPage(nextURL)
		if err != nil {
			return nil, err
		}

		allJobs = append(allJobs, jobs...)

		if nextLink == "" {
			break
		}

		if strings.HasPrefix(nextLink, "/") {
			nextURL = "https://prow.k8s.io" + nextLink
		} else if strings.HasPrefix(nextLink, "?") {
			nextURL = baseURL + nextLink
		} else {
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

	re := regexp.MustCompile(`var allBuilds = (\[.*?\]);`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, "", fmt.Errorf("could not find allBuilds JSON in response")
	}

	var jobs []ProwJob
	if err := json.Unmarshal(matches[1], &jobs); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	nextLinkRe := regexp.MustCompile(`<a href="([^"]+)"[^>]*>&lt;- Older Runs</a>`)
	linkMatches := nextLinkRe.FindSubmatch(body)
	var nextLink string
	if len(linkMatches) >= 2 {
		nextLink = string(linkMatches[1])
	}

	return jobs, nextLink, nil
}

// FormatDuration converts nanoseconds to a human-readable string
func FormatDuration(ns int64) string {
	d := time.Duration(ns)
	return d.Round(time.Second).String()
}
