package vpsadmin

import (
	"errors"
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

type Config struct {
	client *client.Client
}

func (c *Config) getClient() *client.Client {
	return c.client
}

func (c *Config) testAuthentication() error {
	resp, err := c.client.User.Current.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return errors.New(resp.Message)
	}

	return nil
}

func configureClient(apiUrl string, authToken string) (*Config, error) {
	c := client.New(apiUrl)
	c.SetExistingTokenAuth(authToken)
	return &Config{client: c}, nil
}

func waitForOperation(watcher client.BlockingOperationWatcher) error {
	if watcher.IsBlocking() {
		for i := 0; i < 60; i++ {
			resp, err := watcher.WaitForOperation(60)

			if err != nil {
				return err
			}

			if !resp.Status {
				return fmt.Errorf(
					"Error while waiting for operation to finish: %s",
					resp.Message,
				)
			}

			if resp.Output.Finished {
				if resp.Output.Status {
					return nil
				} else {
					return fmt.Errorf("Operation failed")
				}
			}
		}
	}

	return nil
}
