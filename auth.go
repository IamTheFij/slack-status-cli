package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var (
	// These are set via build flags but can be overridden via environment variables.
	defaultClientID     = ""
	defaultClientSecret = ""

	//go:embed "certs/cert.pem"
	certPem []byte
	//go:embed "certs/key.pem"
	keyPem []byte
)

const (
	httpReadTimeout  = 5 * time.Second
	httpWriteTimeout = 10 * time.Second
	httpIdleTimeout  = 120 * time.Second
)

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

	certPath, err := getConfigFilePath("cert.pem")
	if err != nil {
		return "", fmt.Errorf("failed checking config path for cert: %w", err)
	}

	keyPath, err := getConfigFilePath("key.pem")
	if err != nil {
		return "", fmt.Errorf("failed checking config path for key: %w", err)
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// If config files don't exist, use embedded
	if !fileExists(certPath) && !fileExists(keyPath) {
		cert, err := tls.X509KeyPair(certPem, keyPem)
		if err != nil {
			return "", fmt.Errorf("failed loading embedded key pair: %w", err)
		}

		tlsCfg.Certificates = make([]tls.Certificate, 1)
		tlsCfg.Certificates[0] = cert

		// Empty out paths since they don't exist so embeded certs will be used
		certPath = ""
		keyPath = ""
	}

	// Also, should generate TLS certificate to use since https is a required scheme
	server := http.Server{
		Addr:         app.listenHost,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		IdleTimeout:  httpIdleTimeout,
		TLSConfig:    tlsCfg,
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

	if err := server.ListenAndServeTLS(certPath, keyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return "", err
	}

	return code, nil
}

func authenticate() (string, error) {
	app := slackApp{
		userScopes:   []string{"dnd:write", "users.profile:write", "team:read"},
		scopes:       []string{"dnd:write", "users.profile:write", "team:read"},
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
