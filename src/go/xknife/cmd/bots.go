package cmd

import (
	"context"
	"fmt"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/resources"
	"github.com/michimani/gotwi/user/follow"
	ftypes "github.com/michimani/gotwi/user/follow/types"
	"github.com/michimani/gotwi/user/userlookup"
	utypes "github.com/michimani/gotwi/user/userlookup/types"
	"github.com/spf13/cobra"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xknife/pkg"
)

var userCache *pkg.UserCache

func init() {
	rootCmd.AddCommand(getUserCmd, getFollowersCmd)
	// Initialize cache in user's home directory
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".xknife", "cache")
	userCache = pkg.NewUserCache(cacheDir)
}

var getUserCmd = &cobra.Command{
	Use:   "get",
	Short: "Get user",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := getUser(userName)
		if err != nil {
			fmt.Println("Could not get user", userName, err)
			return err
		}
		printUser(output.Data)
		return nil
	},
}

var getFollowersCmd = &cobra.Command{
	Use:   "followers",
	Short: "Check recent followers on Twitter",
	Long: `Check recent followers on Twitter.

⚠️  IMPORTANT: This command requires Twitter API v2 Project-level access.
If you get a 403 Forbidden error, you need to:
1. Create a Project in Twitter Developer Portal
2. Attach your app to the project  
3. Regenerate your API keys from within the project

For bot detection, try using 'get' command instead, which works with basic API access.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureUserId(); err != nil {
			return err
		}
		fmt.Printf("Getting followers for %s (%s)...\n", userName, userId)
		
		followers, err := getFollowers(userId, pageSize)
		if err != nil {
			// Check if it's the Project attachment error
			if strings.Contains(err.Error(), "403") && strings.Contains(err.Error(), "Project") {
				fmt.Println("\n🚨 ERROR: Twitter API v2 Project Required!")
				fmt.Println("This endpoint requires your app to be attached to a Project.")
				fmt.Println("\nTo fix this:")
				fmt.Println("1. Go to https://developer.twitter.com/en/portal/dashboard")
				fmt.Println("2. Create a Project or use an existing one")
				fmt.Println("3. Add your app to the project")
				fmt.Println("4. Regenerate your API keys from within the project")
				fmt.Println("5. Update your .env file with the new keys")
				fmt.Println("\n💡 Alternative: Use 'xknife get --user username' for individual analysis")
				return fmt.Errorf("Twitter API Project setup required")
			}
			return err
		}
		
		for _, user := range followers {
			printUser(user)
		}
		return nil
	},
}

func ensureUserId() error {
	if len(userId) == 0 {
		u, err := getUser(userName)
		if err != nil {
			return err
		}
		userId = gotwi.StringValue(u.Data.ID)
	}
	return nil
}

func getUser(name string) (*utypes.GetByUsernameOutput, error) {
	// Check cache first
	if cachedUser, found := userCache.GetUser(name); found {
		fmt.Printf("[CACHE HIT] Found user %s in cache\n", name)
		return &utypes.GetByUsernameOutput{Data: *cachedUser}, nil
	}

	fmt.Printf("[API CALL] Fetching user %s from API (watch rate limits!)\n", name)
	p := &utypes.GetByUsernameInput{
		Username: name,
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
			fields.UserFieldVerified,
			// fields.UserFieldVerifiedType, // This might not be available in gotwi yet
			fields.UserFieldProtected,
			fields.UserFieldLocation,
			fields.UserFieldPublicMetrics,
		},
	}

	result, err := userlookup.GetByUsername(context.Background(), xClient, p)
	if err != nil {
		return nil, err
	}

	// Cache the result
	userCache.SetUser(name, &result.Data)

	return result, nil
}

func getFollowers(id string, size int) ([]resources.User, error) {
	// Check cache first
	if cachedFollowers, found := userCache.GetFollowers(id, size); found {
		fmt.Printf("[CACHE HIT] Found %d followers in cache\n", len(cachedFollowers))
		return cachedFollowers, nil
	}

	fmt.Printf("[API CALL] Fetching %d followers from API (EXPENSIVE!)\n", size)
	fmt.Println("WARNING: This uses your precious rate limit quota!")
	
	p := &ftypes.ListFollowersInput{
		ID:              id,
		MaxResults:      ftypes.ListMaxResults(size),
		PaginationToken: "",
		Expansions:      nil,
		TweetFields:     nil,
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
			fields.UserFieldVerified,
			fields.UserFieldProtected,
			fields.UserFieldPublicMetrics,
		},
	}

	output, err := follow.ListFollowers(context.Background(), xClient, p)
	if err != nil {
		return nil, err
	}

	// Cache the results
	if output.Data != nil {
		userCache.SetFollowers(id, size, output.Data)
	}

	return output.Data, nil
}

func printUser(u resources.User) {
	fmt.Println("ID:          ", gotwi.StringValue(u.ID))
	fmt.Println("Name:        ", gotwi.StringValue(u.Name))
	fmt.Println("Username:    ", gotwi.StringValue(u.Username))
	fmt.Println("CreatedAt:   ", gotwi.TimeValue(u.CreatedAt))
	
	fmt.Println("Verified:    ", gotwi.BoolValue(u.Verified))
	fmt.Println("Protected:   ", gotwi.BoolValue(u.Protected))
	fmt.Println("Location:    ", gotwi.StringValue(u.Location))
	if u.PublicMetrics != nil {
		m := *u.PublicMetrics
		fmt.Printf("Following: %d, Followers: %d, Tweets: %d, Lists: %d\n",
			gotwi.IntValue(m.FollowingCount), gotwi.IntValue(m.FollowersCount),
			gotwi.IntValue(m.TweetCount), gotwi.IntValue(m.ListedCount),
		)
	}
	fmt.Printf("Score:       %.2f%%", score(u))
	
	// Add scoring breakdown for debugging (optional verbose mode)
	if pageSize <= 5 { // Only show details for small queries
		fmt.Printf(" (Age: %.0fd, Tweets/day: %.1f)", 
			time.Since(gotwi.TimeValue(u.CreatedAt)).Hours()/24.0,
			float64(gotwi.IntValue(u.PublicMetrics.TweetCount))/(time.Since(gotwi.TimeValue(u.CreatedAt)).Hours()/24.0))
	}
	fmt.Println()
}

func score(u resources.User) float64 {
	// https://developer.x.com/en/docs/x-api/data-dictionary/object-model/user
	s := 100.0
	
	// Note: Modern verification is less reliable (paid verification since 2022)
	// Verified accounts still get a bonus, but not automatic 100%
	verifiedBonus := 0.0
	if gotwi.BoolValue(u.Verified) {
		verifiedBonus = 20.0 // Bonus points, not automatic pass
	}
	
	// Protected accounts get a small bonus but aren't automatically human
	protectedBonus := 0.0
	if gotwi.BoolValue(u.Protected) {
		protectedBonus = 10.0 // Small bonus for privacy-conscious behavior
	}
	// 2. follower/following ratio
	// x: 10, 5   , 1  , 0.5 , 0.1, 0.01
	// y:  1, 0.95, 0.9, 0.85, 0.3,    0
	following := float64(gotwi.IntValue(u.PublicMetrics.FollowingCount))
	followers := float64(gotwi.IntValue(u.PublicMetrics.FollowersCount))
	x := followers / math.Max(following, 1.0)
	y := 2 * math.Atan(5*x) / 3 // <-- https://www.geogebra.org/graphing
	s *= y
	// - account duration compared to absolute following (bots are quick to follow many accounts, following more than 20/day avg is usually suspicious)
	// Cutoff date: March 2006, cf.: https://www.oldest.org/technology/oldest-twitter-accounts/
	cutOffDate := time.Date(2006, time.March, 21, 0, 0, 0, 0, time.UTC)
	days := gotwi.TimeValue(u.CreatedAt).Sub(cutOffDate).Hours() / 24.0
	x = math.Min(following/days, 40)
	// x: 1, 5   , 10  , 20 , 30 , 40
	// y: 1, 0.95, 0.85, 0.5, 0.1,  0
	y = (1 + math.Cos(x/(4*math.Pi))) / 2
	s *= y
	
	// Apply bonuses after penalties
	s = math.Min(100, s + verifiedBonus + protectedBonus)
	
	// 3. Account age analysis - very new accounts are often suspicious
	accountAge := time.Since(gotwi.TimeValue(u.CreatedAt))
	ageDays := accountAge.Hours() / 24.0
	
	if ageDays < 30 { // Less than 30 days old
		ageMultiplier := ageDays / 30.0 // 0.0 to 1.0 scale
		s *= (0.3 + 0.7*ageMultiplier) // 30% to 100% based on age
	}
	
	// 4. Tweet frequency analysis - inhuman posting patterns
	tweetCount := float64(gotwi.IntValue(u.PublicMetrics.TweetCount))
	tweetsPerDay := tweetCount / math.Max(ageDays, 1.0)
	
	if tweetsPerDay > 50 { // More than 50 tweets per day average
		// Penalize excessive posting (bots often spam)
		excessFactor := math.Min(tweetsPerDay/50.0, 5.0) // Cap at 5x penalty
		s *= 1.0 / (1.0 + 0.2*(excessFactor-1.0)) // Gradual penalty
	} else if tweetsPerDay < 0.01 && ageDays > 365 { // Very old account with almost no tweets
		s *= 0.7 // Suspicious: old account, no activity
	}
	
	// 5. Absolute follower/following number analysis
	// (reusing followers and following variables from earlier)
	
	// Suspicious pattern: Following many, followers few (typical bot pattern)
	if following > 2000 && followers < 50 {
		s *= 0.4 // Heavy penalty for obvious bot pattern
	}
	
	// Suspicious pattern: Both very low for old account
	if followers < 5 && following < 5 && ageDays > 365 {
		s *= 0.6 // Old account with no social activity
	}
	
	// Suspicious pattern: Following exactly 0 (often indicates suspended/restricted bots)
	if following == 0 && followers > 100 {
		s *= 0.5 // Weird pattern: many followers, following none
	}
	
	// 6. Username patterns (basic heuristics with available data)
	// 6. Username patterns (basic heuristics with available data)
	username := gotwi.StringValue(u.Username)
	if len(username) > 15 && containsExcessiveNumbers(username) {
		s *= 0.8 // Penalty for very long usernames with many numbers
	}
	
	// Additional username patterns
	if hasTypicalBotUsernamePattern(username) {
		s *= 0.7 // Penalty for bot-like username patterns
	}
	
	// 7. Account name vs username consistency (names that are just usernames often indicate bots)
	name := gotwi.StringValue(u.Name)
	if len(name) > 0 && strings.ToLower(name) == strings.ToLower(username) {
		s *= 0.9 // Small penalty for identical name/username (lazy bot behavior)
	}
	
	// 8. TODO: Add more sophisticated checks when API provides:
	// - Profile picture analysis (default avatars)
	// - Bio analysis (empty, repetitive, keyword stuffing)
	// - Tweet content analysis (frequency, similarity, hashtag abuse)
	
	return math.Min(100, math.Max(0, s))
}

// containsExcessiveNumbers checks if username has >30% digits (common bot pattern)
func containsExcessiveNumbers(username string) bool {
	if len(username) == 0 {
		return false
	}
	
	digitCount := 0
	for _, r := range username {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}
	
	return float64(digitCount)/float64(len(username)) > 0.3
}

// hasTypicalBotUsernamePattern detects common bot username patterns
func hasTypicalBotUsernamePattern(username string) bool {
	if len(username) == 0 {
		return false
	}
	
	username = strings.ToLower(username)
	
	// Pattern 1: Ends with many consecutive numbers (e.g., "user12345678")
	if len(username) > 6 {
		lastPart := username[len(username)-6:]
		digitCount := 0
		for _, r := range lastPart {
			if r >= '0' && r <= '9' {
				digitCount++
			}
		}
		if digitCount >= 5 { // Last 6 chars have 5+ digits
			return true
		}
	}
	
	// Pattern 2: Alternating letters and numbers in a pattern (e.g., "a1b2c3d4e5")
	if len(username) >= 8 {
		alternatingCount := 0
		for i := 0; i < len(username)-1; i++ {
			curr := username[i]
			next := username[i+1]
			isCurrentDigit := curr >= '0' && curr <= '9'
			isNextDigit := next >= '0' && next <= '9'
			if isCurrentDigit != isNextDigit { // Different types (letter vs digit)
				alternatingCount++
			}
		}
		// If more than 60% of adjacent pairs alternate, it's suspicious
		if float64(alternatingCount)/float64(len(username)-1) > 0.6 {
			return true
		}
	}
	
	// Pattern 3: Common bot prefixes/suffixes
	botPatterns := []string{
		"bot", "auto", "fake", "spam", "test", "temp",
		"user", "account", "profile", "real", "official",
	}
	
	for _, pattern := range botPatterns {
		if strings.Contains(username, pattern) && len(username) > len(pattern)+3 {
			// Contains bot-like word and has additional characters (likely numbers)
			nonPatternChars := strings.ReplaceAll(username, pattern, "")
			if containsExcessiveNumbers(nonPatternChars) {
				return true
			}
		}
	}
	
	return false
}
