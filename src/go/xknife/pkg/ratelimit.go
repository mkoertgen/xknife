package pkg

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// RateLimitInfo holds rate limit information from API response headers
type RateLimitInfo struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// ParseRateLimitHeaders extracts rate limit info from HTTP headers
func ParseRateLimitHeaders(headers http.Header) *RateLimitInfo {
	info := &RateLimitInfo{}
	
	if limit := headers.Get("x-rate-limit-limit"); limit != "" {
		info.Limit, _ = strconv.Atoi(limit)
	}
	
	if remaining := headers.Get("x-rate-limit-remaining"); remaining != "" {
		info.Remaining, _ = strconv.Atoi(remaining)
	}
	
	if reset := headers.Get("x-rate-limit-reset"); reset != "" {
		if resetTime, err := strconv.ParseInt(reset, 10, 64); err == nil {
			info.Reset = time.Unix(resetTime, 0)
		}
	}
	
	return info
}

// WaitForReset sleeps until rate limit resets
func (r *RateLimitInfo) WaitForReset() {
	if r.Remaining <= 0 && !r.Reset.IsZero() {
		waitTime := time.Until(r.Reset)
		if waitTime > 0 {
			fmt.Printf("Rate limit exceeded. Waiting %v until reset at %v\n", 
				waitTime.Round(time.Second), r.Reset.Format(time.RFC3339))
			time.Sleep(waitTime)
		}
	}
}

// ShouldWait returns true if we should wait before making another request
func (r *RateLimitInfo) ShouldWait(threshold int) bool {
	return r.Remaining <= threshold
}