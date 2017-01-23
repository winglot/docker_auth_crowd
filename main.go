package main

import (
	"bufio"
	"github.com/jessevdk/go-flags"
	"go.jona.me/crowd"
	"os"
	"strings"
)

const (
	AuthAllowed = 0
	AuthDenied  = 1
	AuthNoMatch = 2
	AuthError   = 3
)

type credentials struct {
	username string
	password string
}

var options struct {
	Crowd    string `long:"crowd" description:"URL to Crowd Server" required:"1"`
	AppName  string `long:"app-name" description:"Application name in Crowd" required:"1"`
	Password string `long:"app-password" description:"Password for application in Crowd" required:"1"`
}

func readCredentials() credentials {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	splitted := strings.Split(text, " ")
	if len(splitted) != 2 {
		os.Exit(AuthError)
	}

	return credentials{
		username: strings.TrimSpace(splitted[0]),
		password: strings.TrimSpace(splitted[1]),
	}

}

func accessDenied(err error) bool {
	// From: https://developer.atlassian.com/display/CROWDDEV/Using+the+Crowd+REST+APIs#UsingtheCrowdRESTAPIs-HTTPResponseCodesandErrorResponses
	deniedMessages := [...]string{
		"APPLICATION_ACCESS_DENIED",
		"EXPIRED_CREDENTIAL",
		"INACTIVE_ACCOUNT",
		"INVALID_USER_AUTHENTICATION",
		"INVALID_CREDENTIAL",
		"INVALID_EMAIL",
		"INVALID_USER",
		"USER_NOT_FOUND",
	}

	for _, msg := range deniedMessages {
		if msg == err.Error() {
			return true
		}
	}

	return false
}

func tryAuthenticate(userCreds credentials, crowdUrl, appName, appPassword string) (bool, error) {
	client, err := crowd.New(appName, appPassword, crowdUrl)
	if err != nil {
		return false, err
	}

	_, err = client.Authenticate(userCreds.username, userCreds.password)
	if err != nil {
		if accessDenied(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func main() {
	if _, err := flags.Parse(&options); err != nil {
		os.Exit(AuthError)
	}

	auth := readCredentials()

	authenticated, err := tryAuthenticate(auth, options.Crowd, options.AppName, options.Password)
	if err != nil {
		os.Exit(AuthError)
	}

	if authenticated {
		os.Exit(AuthAllowed)
	}

	os.Exit(AuthDenied)
}
