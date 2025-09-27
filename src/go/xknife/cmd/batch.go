package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/michimani/gotwi"
	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Batch operations for rate limit optimization",
}

var analyzeFileCmd = &cobra.Command{
	Use:   "analyze-file [file]",
	Short: "Analyze user data from a JSON file (offline, no API calls)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return analyzeUsersFromFile(args[0])
	},
}

var collectUsersCmd = &cobra.Command{
	Use:   "collect-users [usernames-file]",
	Short: "Collect user data with rate limit awareness (saves to JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return collectUsersFromFile(args[0])
	},
}

var clearCacheCmd = &cobra.Command{
	Use:   "clear-cache",
	Short: "Clear the local cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, _ := os.UserHomeDir()
		cacheDir := filepath.Join(homeDir, ".xknife", "cache")
		
		err := os.RemoveAll(cacheDir)
		if err != nil {
			return err
		}
		
		fmt.Println("Cache cleared successfully")
		return nil
	},
}

func init() {
	batchCmd.AddCommand(analyzeFileCmd, collectUsersCmd, clearCacheCmd)
	rootCmd.AddCommand(batchCmd)
}

func analyzeUsersFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	fmt.Printf("Analyzing %d users from %s (offline analysis):\n\n", len(users), filename)

	botCount := 0
	for i, userData := range users {
		// This is a simplified analysis - you'd need to reconstruct the User struct
		username := userData["username"].(string)
		verified := userData["verified"].(bool)
		
		// Simplified bot scoring
		score := 50.0 // Base score
		if verified {
			score = 100.0
		}
		
		fmt.Printf("%d. %s - Score: %.1f%%", i+1, username, score)
		if score < 50 {
			fmt.Print(" [LIKELY BOT]")
			botCount++
		}
		fmt.Println()
	}
	
	fmt.Printf("\nSummary: %d/%d accounts flagged as likely bots (%.1f%%)\n", 
		botCount, len(users), float64(botCount)/float64(len(users))*100)

	return nil
}

func collectUsersFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var usernames []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		username := strings.TrimSpace(scanner.Text())
		if username != "" && !strings.HasPrefix(username, "#") {
			usernames = append(usernames, username)
		}
	}

	if len(usernames) == 0 {
		return fmt.Errorf("no usernames found in file")
	}

	fmt.Printf("Found %d usernames. Starting collection with rate limit awareness...\n", len(usernames))
	fmt.Println("This will be VERY slow due to API rate limits!")

	var collectedUsers []interface{}
	
	for i, username := range usernames {
		fmt.Printf("\n[%d/%d] Processing: %s\n", i+1, len(usernames), username)
		
		// Use existing getUser function which has caching
		output, err := getUser(username)
		if err != nil {
			fmt.Printf("ERROR: Failed to get user %s: %v\n", username, err)
			
			// Check if it's a rate limit error
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
				fmt.Println("RATE LIMITED! Waiting 15 minutes...")
				time.Sleep(15 * time.Minute)
				
				// Retry once
				output, err = getUser(username)
				if err != nil {
					fmt.Printf("Still failed after wait: %v\n", err)
					continue
				}
			} else {
				continue
			}
		}

		if output != nil && len(gotwi.StringValue(output.Data.ID)) > 0 {
			collectedUsers = append(collectedUsers, output.Data)
		}

		// Small delay between requests to be nice to the API
		if i < len(usernames)-1 {
			fmt.Println("Waiting 2 seconds before next request...")
			time.Sleep(2 * time.Second)
		}
	}

	// Save collected data
	outputFile := fmt.Sprintf("users_data_%s.json", time.Now().Format("20060102_150405"))
	jsonData, err := json.MarshalIndent(collectedUsers, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	err = ioutil.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	fmt.Printf("\n✅ Successfully collected %d users and saved to %s\n", len(collectedUsers), outputFile)
	fmt.Printf("You can now analyze this data offline using: xknife batch analyze-file %s\n", outputFile)

	return nil
}