package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

// statusInfo contains all args passed from the command line
type statusInfo struct {
	emoji, statusText string
	duration          time.Duration
	snooze            bool
	accessToken       string
}

// getExipirationTime returns epoch time that status should expire from the duration.
func (si statusInfo) getExpirationTime() int64 {
	if si.duration == 0 {
		return 0
	}

	return time.Now().Add(si.duration).Unix()
}

// getConfigFilePath returns the path of a given file within the config folder.
// The config folder will be created in ~/.local/config/slack-status-cli if it does not exist.
func getConfigFilePath(filename string) string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = "~/.local/config"
	}

	configDir := filepath.Join(configHome, "slack-status-cli")
	_ = os.MkdirAll(configDir, 0755)

	return filepath.Join(configDir, filename)
}

// readWriteAccessToken will store and retrieve access tokens for future use.
func readWriteAccessToken(accessToken string) (string, error) {
	tokenFile := getConfigFilePath("token")

	if accessToken != "" {
		err := ioutil.WriteFile(tokenFile, []byte(accessToken), 0600)
		if err != nil {
			fmt.Println("Error writing access token")
		}

		return accessToken, err
	}

	content, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		fmt.Println("No token provided on command line or in file")

		return "", err
	}

	return string(content), nil
}

// createClient will return a Slack client with the provided access token.
// If that token is empty, it will try to get one from the config folder.
func createClient(accessToken string) (*slack.Client, error) {
	var err error

	accessToken, err = readWriteAccessToken(accessToken)
	if err != nil {
		fmt.Println("error reading access token")

		return nil, err
	}

	client := slack.New(accessToken)

	return client, nil
}

// readDurationArgs will attempt to find a duration within command line args rather than flags.
// It will look for a prefixed duration. eg. "5m :cowboy: Howdy y'all" and a postfix duration
// following the word "for". eg. ":dancing: Dancing for 1h".
func readDurationArgs(args []string) ([]string, *time.Duration) {
	// If there are no args, we have no duration
	if len(args) == 0 {
		return args, nil
	}

	// Try to parse the first value
	durationVal, err := time.ParseDuration(args[0])
	if err == nil {
		// Found a duration, return the trimmed args and duration
		return args[1:], &durationVal
	}

	// If the args are less than two, then we don't have a "for <duration>" expression
	minArgsForSuffix := 2
	if len(args) < minArgsForSuffix {
		return args, nil
	}

	// Check for a "for <duration>" expression at end of args
	if strings.ToLower(args[len(args)-2]) == "for" {
		durationVal, err = time.ParseDuration(args[len(args)-1])
		if err == nil {
			// Found a duration, return the trimmed args and duration
			return args[:len(args)-2], &durationVal
		}
	}

	// Default return input
	return args, nil
}

// readFlags will read all flags off the command line.
func readFlags() statusInfo {
	snooze := flag.Bool("snooze", false, "snooze notifications")
	duration := flag.Duration("duration", 0, "duration to set status for")
	emoji := flag.String("emoji", "", "emoji to use as status")
	accessToken := flag.String("access-token", "", "slack access token")

	flag.Parse()

	// Freeform input checks the first argument to see if it's a duration
	args := flag.Args()

	// Duration was not set via a flag, check the args
	if *duration == 0 {
		var parsedDuration *time.Duration
		args, parsedDuration = readDurationArgs(args)

		if parsedDuration != nil {
			duration = parsedDuration
		}
	}

	if *emoji == "" && len(args) > 0 {
		if args[0][0] == ':' && args[0][len(args[0])-1] == ':' {
			emoji = &args[0]
			args = args[1:]
		}
	}

	statusText := strings.Join(args, " ")

	return statusInfo{
		duration:    *duration,
		snooze:      *snooze,
		emoji:       *emoji,
		accessToken: *accessToken,
		statusText:  statusText,
	}
}

func main() {
	args := readFlags()

	client, err := createClient(args.accessToken)
	if err != nil {
		fmt.Println("error getting client")
		panic(err)
	}

	err = client.SetUserCustomStatus(args.statusText, args.emoji, args.getExpirationTime())
	if err != nil {
		fmt.Println("error setting status")
		panic(err)
	}

	if args.snooze {
		_, err = client.SetSnooze(int(args.duration.Minutes()))
		if err != nil {
			fmt.Println("error setting snooze")
			panic(err)
		}
	} else {
		_, err = client.EndSnooze()
		if err != nil && err.Error() != "snooze_not_active" {
			fmt.Println("error ending snooze")
			panic(err)
		}
	}
}
