package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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
func FetchJobHistory(jobName string) ([]ProwJob, error) {
	// Construct URL - probing generic GS bucket first
	// We might need to support other buckets if 404, but standard is kubernetes-ci-logs
	url := fmt.Sprintf("https://prow.k8s.io/job-history/gs/kubernetes-ci-logs/logs/%s", jobName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// Regex to find the JSON array
	// var allBuilds = [...];
	re := regexp.MustCompile(`var allBuilds = (\[.*?\]);`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find allBuilds JSON in response")
	}

	var jobs []ProwJob
	if err := json.Unmarshal(matches[1], &jobs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return jobs, nil
}

// FormatDuration converts nanoseconds to a human-readable string
func FormatDuration(ns int64) string {
	d := time.Duration(ns)
	return d.Round(time.Second).String()
}
