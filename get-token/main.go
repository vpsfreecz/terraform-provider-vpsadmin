package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/vpsfreecz/vpsadmin-go-client/client"
	"golang.org/x/crypto/ssh/terminal"
)

type options struct {
	apiUrl string
	lifetime string
	interval int
	tfvars string
	username string
	password string
}

func main() {
	opts := parseOptions()

	if err := getCredentials(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		return
	}

	token, err := getToken(opts)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		return
	}

	if err := writeToken(token, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		return
	}
}

func parseOptions() *options {
	opts := &options{}

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage:\n  %s [options] [api url]\n\nOptions:\n", os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.StringVar(
		&opts.lifetime,
		"lifetime",
		"renewable_auto",
		"Token lifetime",
	)

	flag.IntVar(
		&opts.interval,
		"interval",
		3600,
		"How long should the token be valid, in seconds",
	)

	flag.StringVar(
		&opts.username,
		"user",
		"",
		"User name",
	)

	flag.StringVar(
		&opts.tfvars,
		"tfvars",
		"",
		"Write the token to a dedicated .tfvars file",
	)

	flag.Parse()

	args := flag.Args()

	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "Error: too many arguments")
		flag.PrintDefaults()
	} else if len(args) == 1 {
		opts.apiUrl = args[0]
	} else {
		opts.apiUrl = "https://api.vpsfree.cz"
	}

	return opts
}

func getCredentials(opts *options) error {
	reader := bufio.NewReader(os.Stdin)

	if opts.username == "" {
	    fmt.Print("Username: ")
		username, err := reader.ReadString('\n')

		if err != nil {
			return err
		}

		opts.username = strings.TrimSpace(username)
	}

    fmt.Print("Password: ")
    bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")

	if err != nil {
		return err
    }

	opts.password = strings.TrimSpace(string(bytePassword))
	return nil
}

func getToken(opts *options) (string, error) {
	api := client.New(opts.apiUrl)
	err := api.SetNewTokenAuth(
		opts.username,
		opts.password,
		opts.lifetime,
		opts.interval,
	)

	if err != nil {
		return "", err
	}

	return api.Authentication.(*client.TokenAuth).Token, nil
}

func writeToken(token string, opts *options) error {
	if opts.tfvars == "" {
		fmt.Printf(token)
		return nil
	}

	f, err := os.Create(opts.tfvars)

	if err != nil {
		return err
	}

	defer f.Close()
	fmt.Fprintf(f, "vpsadmin_token = \"%s\"", token)
	return nil
}
