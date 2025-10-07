package usecase

import "time"

// stringPtr is a helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}

// intPtr is a helper function to convert int to *int
func intPtr(i int) *int {
	return &i
}

// timePtr is a helper function to convert time.Time to *time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
