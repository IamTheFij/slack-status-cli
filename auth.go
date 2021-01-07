package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var (
	// These are set via build flags but can be overriden via environment variables.
	defaultClientID     = ""
	defaultClientSecret = ""
)

func getEnvOrDefault(name, defaultValue string) string {
	val, ok := os.LookupEnv(name)
	if ok {
		return val
	}

	return defaultValue
}

type slackApp struct {
	clientID, clientSecret, redirectURI string
	scopes, userScopes                  []string
	listenHost, listenPath              string
}

func (app slackApp) getAuthURL() string {
	scopes := strings.Join(app.scopes, ",")
	userScopes := strings.Join(app.userScopes, ",")

	return fmt.Sprintf(
		"https://slack.com/oauth/authorize?scope=%s&user_scope=%s&client_id=%s&redirect_uri=%s",
		scopes,
		userScopes,
		app.clientID,
		app.redirectURI,
	)
}

func (app slackApp) listenForCode() (string, error) {
	// start an http listener and listen for the redirect and return the code from params
	var code string

	// Also, should generate TLS certificate to use since https is a required scheme
	server := http.Server{
		Addr:         app.listenHost,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	http.HandleFunc(app.listenPath, func(w http.ResponseWriter, r *http.Request) {
		codes := r.URL.Query()["code"]
		if len(codes) == 0 {
			log.Fatal("no oauth code found in response")
		}

		code = codes[0]
		fmt.Fprintf(w, "Got code %s", code)

		// Shutdown after response
		go func() {
			if err := server.Shutdown(context.Background()); err != nil {
				fmt.Println("Fatal?")
				log.Fatal(err)
			}
		}()
	})

	certPath := getConfigFilePath("cert.pem")
	keyPath := getConfigFilePath("key.pem")

	if !fileExists(certPath) || !fileExists(keyPath) {
		if err := generateSelfSignedCertificates(certPath, keyPath); err != nil {
			return "", err
		}
	}

	if err := server.ListenAndServeTLS(certPath, keyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return "", err
	}

	return code, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func generateSelfSignedCertificates(certPath, keyPath string) error {
	command := exec.Command(
		"openssl",
		"req",
		"-x509",
		"-subj",
		"/C=US/O=Slack Status CLI/CN=localhost:8888",
		"-nodes",
		"-days",
		"365",
		"-newkey",
		"rsa:2048",
		"-keyout",
		keyPath,
		"-out",
		certPath,
	)

	return command.Run()
}

func authenticate() (string, error) {
	app := slackApp{
		userScopes:   []string{"dnd:write", "users.profile:write"},
		scopes:       []string{"dnd:write", "users.profile:write"},
		clientID:     getEnvOrDefault("CLIENT_ID", defaultClientID),
		clientSecret: getEnvOrDefault("CLIENT_SECRET", defaultClientSecret),
		redirectURI:  "https://localhost:8888/auth",
		listenHost:   "localhost:8888",
		listenPath:   "/auth",
	}

	fmt.Println("To authenticate, go to the following URL:")
	fmt.Println("NOTE: After you authenticate with Slack, it will redirect you to a server running on your local computer. Your browser will present a security error because it cann't verify the server. You will need to manually add an exception or tell your browser to proceed anyway.")
	fmt.Println(app.getAuthURL())

	code, err := app.listenForCode()
	if err != nil {
		return "", err
	}

	accessToken, _, err := slack.GetOAuthToken(&http.Client{}, app.clientID, app.clientSecret, code, app.redirectURI)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
