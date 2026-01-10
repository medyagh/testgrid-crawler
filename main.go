package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"test-grid-crawler/pkg/crawler"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-grid-crawler <job-name>")
		fmt.Println("Example: test-grid-crawler minikube-periodics#ci-minikube-integration")
		fmt.Println("         test-grid-crawler ci-minikube-integration")
		os.Exit(1)
	}

	input := os.Args[1]
	jobName := parseJobName(input)

	fmt.Printf("Fetching history for job: %s\n", jobName)

	c := crawler.New(crawler.Config{
		JobName:  jobName,
		MaxPages: 2, // Configurable limit
	})

	jobs, err := c.Run()
	if err != nil {
		fmt.Printf("Error scraping job history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d jobs.\n\n", len(jobs))

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "JOB ID\tSTATUS\tSTARTED\tDURATION\tURL")

	for _, job := range jobs {
		url := job.SpyglassLink
		if strings.HasPrefix(url, "/") {
			url = "https://prow.k8s.io" + url
		}

		started := job.Started
		t, err := time.Parse(time.RFC3339, job.Started)
		if err == nil {
			started = t.Format("2006-01-02 15:04")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", job.ID, job.Result, started, crawler.FormatDuration(job.Duration), url)
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
