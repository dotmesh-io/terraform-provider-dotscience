package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dotmesh-io/terraform-provider-dotscience/pkg/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
)

// Config represents the Linode provider configuration
type Client struct {
	URL      string
	Username string
	Password string
}

// Client returns a fully initialized Linode client
func (client *Client) Request(method, endpoint string, data io.Reader) ([]byte, error) {
	if logging.IsDebugOrHigher() {
		log.Printf("[DEBUG] running api request: %s:%s %s %s%s", client.Username, "<password>", method, client.URL, endpoint)
	}
	url := fmt.Sprintf("%s%s", client.URL, endpoint)
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(client.Username, client.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response, readBodyErr := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if readBodyErr != nil {
			return nil, readBodyErr
		}
		return response, nil
	} else {
		return response, fmt.Errorf("Got non-200 status code %d, for %s %s - original error: %s", resp.StatusCode, method, url, err)
	}
}

func (client *Client) Version() (string, error) {
	data, err := client.Request("GET", "/v2/version", nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (client *Client) ListRunners() ([]types.Runner, error) {
	data, err := client.Request("GET", "/admin/v1/runners", nil)
	if err != nil {
		return nil, err
	}
	runners := &[]types.Runner{}
	if err = json.Unmarshal(data, runners); err != nil {
		return nil, fmt.Errorf("failed decoding ListRunners response, err: %#v\n%#v\n", err, data)
	}
	return *runners, nil
}

func (client *Client) StopRunnerTasks(runner types.Runner) error {
	_, err := client.Request("POST", fmt.Sprintf("/admin/v1/runners/%s/%s/action", runner.AccountID, runner.ID), strings.NewReader(`{"action":"stop_all_tasks"}`))
	return err
}

func (client *Client) DeleteRunner(runner types.Runner) error {
	_, err := client.Request("DELETE", fmt.Sprintf("/admin/v1/runners/%s/%s", runner.AccountID, runner.ID), nil)
	return err
}
