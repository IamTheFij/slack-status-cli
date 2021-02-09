package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var version = "dev"

// statusInfo contains all args passed from the command line
type statusInfo struct {
	emoji, statusText string
	duration          time.Duration
	snooze            bool
}

// commandOptions contains non-status options passed to the command
type commandOptions struct {
	login, makeDefault, showVersion bool
	domain                          string
}

// getExipirationTime returns epoch time that status should expire from the duration.
func (si statusInfo) getExpirationTime() int64 {
	if si.duration == 0 {
		return 0
	}

	return time.Now().Add(si.duration).Unix()
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
func readFlags() (statusInfo, commandOptions) {
	// Non-status flags
	login := flag.Bool("login", false, "login to a Slack workspace")
	domain := flag.String("domain", "", "domain to set status on")
	makeDefault := flag.Bool("make-default", false, "set the current domain to default")
	showVersion := flag.Bool("version", false, "show version and exit")

	// Status flags
	snooze := flag.Bool("snooze", false, "snooze notifications")
	duration := flag.Duration("duration", 0, "duration to set status for")
	emoji := flag.String("emoji", "", "emoji to use as status")

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
			duration:   *duration,
			snooze:     *snooze,
			emoji:      *emoji,
			statusText: statusText,
		}, commandOptions{
			login:       *login,
			domain:      *domain,
			makeDefault: *makeDefault,
			showVersion: *showVersion,
		}
}

// loginAndSave will return a client after a new login flow and save the results
func loginAndSave(domain string) (*slack.Client, error) {
	accessToken, err := authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate new login: %w", err)
	}

	client := slack.New(accessToken)

	if domain == "" {
		info, err := client.GetTeamInfo()
		if err != nil {
			return client, fmt.Errorf("failed to get team info: %w", err)
		}

		domain = info.Domain
	}

	err = saveLogin(domain, accessToken)
	if err != nil {
		return client, fmt.Errorf("failed saving new login info: %w", err)
	}

	return client, err
}

// getClient returns a client either via the provided login or default login
func getClient(domain string) (*slack.Client, error) {
	var accessToken string

	var err error

	if domain == "" {
		accessToken, err = getDefaultLogin()
		if err != nil {
			return nil, fmt.Errorf("failed to get default login: %w", err)
		}
	} else {
		accessToken, err = getLogin(domain)
		if err != nil {
			return nil, fmt.Errorf("failed to get login for domain %s: %w", domain, err)
		}
	}

	return slack.New(accessToken), nil
}

func main() {
	status, options := readFlags()

	if options.showVersion {
		fmt.Println("version:", version)

		return
	}

	var client *slack.Client

	var err error

	// If the new-auth flag is present, force an auth flow
	if options.login {
		client, err = loginAndSave(options.domain)
	} else {
		client, err = getClient(options.domain)
	}

	// We encountered some error in logging in
	if err != nil {
		fmt.Println("Unable to create Slack client. Have you logged in yet? Try using `-login`")
		log.Fatal(fmt.Errorf("failed to get or save client: %w", err))
	}

	// If a domain is provided and asked to make default, save it to config
	if options.makeDefault && options.domain != "" {
		if err = saveDefaultLogin(options.domain); err != nil {
			log.Fatal(fmt.Errorf("failed saving default domain %s: %w", options.domain, err))
		}
	}

	err = client.SetUserCustomStatus(status.statusText, status.emoji, status.getExpirationTime())
	if err != nil {
		fmt.Println("error setting status")
		panic(err)
	}

	if status.snooze {
		_, err = client.SetSnooze(int(status.duration.Minutes()))
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
