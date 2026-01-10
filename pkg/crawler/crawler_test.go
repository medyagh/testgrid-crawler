package crawler

import (
	"reflect"
	"testing"
	"time"
)

func TestFilterJobs(t *testing.T) {
	tests := []struct {
		name     string
		jobs     []ProwJob
		config   Config
		expected []ProwJob
	}{
		{
			name: "No filter",
			jobs: []ProwJob{
				{ID: "1", Result: StatusSuccess, Duration: int64(time.Minute)},
				{ID: "2", Result: StatusFailure, Duration: int64(time.Minute)},
			},
			config: Config{},
			expected: []ProwJob{
				{ID: "1", Result: StatusSuccess, Duration: int64(time.Minute)},
				{ID: "2", Result: StatusFailure, Duration: int64(time.Minute)},
			},
		},
		{
			name: "Skip status",
			jobs: []ProwJob{
				{ID: "1", Result: StatusSuccess, Duration: int64(time.Minute)},
				{ID: "2", Result: StatusFailure, Duration: int64(time.Minute)},
			},
			config: Config{SkipStatuses: []string{StatusSuccess}},
			expected: []ProwJob{
				{ID: "2", Result: StatusFailure, Duration: int64(time.Minute)},
			},
		},
		{
			name: "Skip multiple statuses (case insensitive)",
			jobs: []ProwJob{
				{ID: "1", Result: StatusSuccess, Duration: int64(time.Minute)},
				{ID: "2", Result: StatusFailure, Duration: int64(time.Minute)},
				{ID: "3", Result: StatusAborted, Duration: int64(time.Minute)},
			},
			config: Config{SkipStatuses: []string{"success", "Aborted"}},
			expected: []ProwJob{
				{ID: "2", Result: StatusFailure, Duration: int64(time.Minute)},
			},
		},
		{
			name: "Min duration",
			jobs: []ProwJob{
				{ID: "1", Result: StatusSuccess, Duration: int64(10 * time.Minute)},
				{ID: "2", Result: StatusSuccess, Duration: int64(5 * time.Minute)},
			},
			config: Config{MinDuration: 8 * time.Minute},
			expected: []ProwJob{
				{ID: "1", Result: StatusSuccess, Duration: int64(10 * time.Minute)},
			},
		},
		{
			name: "Combined filter",
			jobs: []ProwJob{
				{ID: "1", Result: StatusSuccess, Duration: int64(10 * time.Minute)},
				{ID: "2", Result: StatusFailure, Duration: int64(2 * time.Minute)},  // Short failure
				{ID: "3", Result: StatusAborted, Duration: int64(10 * time.Minute)}, // Skipped status
			},
			config: Config{
				SkipStatuses: []string{StatusAborted},
				MinDuration:  5 * time.Minute,
			},
			expected: []ProwJob{
				{ID: "1", Result: StatusSuccess, Duration: int64(10 * time.Minute)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterJobs(tt.jobs, tt.config)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("FilterJobs() = %v, want %v", got, tt.expected)
			}
		})
	}
}
