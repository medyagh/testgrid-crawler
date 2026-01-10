package main

import (
	"fmt"
	"os"
	"strings"
	"test-grid-crawler/pkg/crawler"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/pflag"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
)

func main() {
	skipStatus := pflag.String("skip-status", "", "Comma-separated list of statuses to skip (e.g., SUCCESS,Aborted)")
	minDuration := pflag.Duration("min-duration", 0, "Minimum job duration (e.g., 5m, 1h)")
	pflag.Parse()

	args := pflag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: test-grid-crawler [flags] <job-name>")
		fmt.Println("Example: test-grid-crawler -skip-status SUCCESS,Aborted minikube-periodics#ci-minikube-integration")
		fmt.Println("         test-grid-crawler --min-duration 10m ci-minikube-integration")
		fmt.Println("         test-grid-crawler ci-minikube-integration")
		pflag.PrintDefaults()
		os.Exit(1)
	}

	input := args[0]
	jobName := parseJobName(input)

	fmt.Printf("Fetching history for job: %s\n", jobName)

	skipList := []string{}
	if *skipStatus != "" {
		skipList = strings.Split(*skipStatus, ",")
	}

	c := crawler.New(crawler.Config{
		JobName:      jobName,
		MaxPages:     2, // Configurable limit
		SkipStatuses: skipList,
		MinDuration:  *minDuration,
	})

	jobs, err := c.Run()
	if err != nil {
		fmt.Printf("Error scraping job history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d jobs.\n\n", len(jobs))

	// Print table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"JOB ID", "STATUS", "STARTED", "DURATION", "PR", "URL"})
	table.SetBorder(false) // User seems to prefer simple list, but tablewriter default valid too. Let's try No Border to match previous style but aligned.
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	// table.SetAutoWrapText(false) // Ensure URLs don't wrap weirdly if not needed

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

		prRef := "master"
		if job.Refs != nil {
			if len(job.Refs.Pulls) > 0 {
				prRef = fmt.Sprintf("PR %d", job.Refs.Pulls[0].Number)
			} else if job.Refs.BaseRef != "" {
				prRef = job.Refs.BaseRef
			}
		}


		// Determine color
		colorCode := ""
		if job.Result == crawler.StatusFailure {
			colorCode = ColorRed
		} else if job.Result != crawler.StatusSuccess {
			colorCode = ColorYellow
		}

		c := func(s string) string {
			if colorCode == "" {
				return s
			}
			return colorCode + s + ColorReset
		}

		table.Append([]string{
			c(job.ID),
			c(job.Result),
			c(started),
			c(crawler.FormatDuration(job.Duration)),
			c(prRef),
			c(url),
		})
	}
	table.Render()
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
