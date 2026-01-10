package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: prow-crawler <job-name>")
		fmt.Println("Example: prow-crawler minikube-periodics#ci-minikube-integration")
		fmt.Println("         prow-crawler ci-minikube-integration")
		os.Exit(1)
	}

	input := os.Args[1]
	jobName := parseJobName(input)

	fmt.Printf("Fetching history for job: %s\n", jobName)

	jobs, err := FetchJobHistory(jobName)
	if err != nil {
		fmt.Printf("Error scraping job history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d jobs.\n\n", len(jobs))

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "JOB ID\tSTATUS\tSTARTED\tDURATION\tURL")

	// Jobs seem to be sorted by most recent first in the JSON usually, but we can rely on order returned
	// or sort if needed. The snippet showed them descending.
	// We will list all of them or maybe top 20 to avoid spamming?
	// The user asked for "all url for the most recent jobs", "sorted by most recent".
	// Let's print top 50 by default to be safe, or all.
	// The User said "list all url for the most recent jobs".
	// I will print all, but maybe pagination is better?
	// For CLI, piping is easy. I'll print all.

	for _, job := range jobs {
		// Prow URL: https://prow.k8s.io/view/gs/kubernetes-ci-logs/logs/ci-minikube-integration/2009755303301615616
		// The SpyglassLink in JSON is relative: /view/gs/kubernetes-ci-logs/logs/ci-minikube-integration/2009755303301615616
		// We should prepend https://prow.k8s.io if it starts with /
		url := job.SpyglassLink
		if strings.HasPrefix(url, "/") {
			url = "https://prow.k8s.io" + url
		}

		started := job.Started
		// Try to parse time to be nicer?
		// "2026-01-09T22:33:06Z"
		t, err := time.Parse(time.RFC3339, job.Started)
		if err == nil {
			started = t.Format("2006-01-02 15:04")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", job.ID, job.Result, started, FormatDuration(job.Duration), url)
	}
	w.Flush()
}

func parseJobName(input string) string {
	// Handle "dashboard#tab" format
	if strings.Contains(input, "#") {
		parts := strings.Split(input, "#")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	// Also handle full URLs if pasted? Not requested but nice.
	// For now just basic string.
	return input
}
