package gatussdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GetAllSuiteStatuses retrieves the status of all configured suites.
//
// Example:
//
//	statuses, err := client.GetAllSuiteStatuses(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, status := range statuses {
//	    fmt.Printf("Suite: %s (Key: %s)\n", status.Name, status.Key)
//	}
func (c *Client) GetAllSuiteStatuses(ctx context.Context) ([]SuiteStatus, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/suites/statuses")
	if err != nil {
		return nil, err
	}
	var statuses []SuiteStatus
	if err := c.decodeResponse(resp, &statuses); err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetSuiteStatusByKey retrieves the status of a specific suite by its key.
// The key should be in the format: {group}_{name}.
//
// Example:
//
//	status, err := client.GetSuiteStatusByKey(context.Background(), "_check-authentication")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Suite %s has %d results\n", status.Name, len(status.Results))
func (c *Client) GetSuiteStatusByKey(ctx context.Context, key string) (*SuiteStatus, error) {
	if key == "" {
		return nil, &ValidationError{
			Field:   "key",
			Message: "cannot be empty",
		}
	}
	path := fmt.Sprintf("/api/v1/suites/%s/statuses", url.PathEscape(key))
	resp, err := c.doRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	var status SuiteStatus
	if err := c.decodeResponse(resp, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// GetSuiteStatus retrieves the status of a specific suite by its group and name.
// The key is generated internally using GenerateKey.
//
// Example:
//
//	status, err := client.GetSuiteStatus(context.Background(), "", "check-authentication")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, result := range status.Results {
//	    fmt.Printf("Suite execution at %s: success=%v, duration=%dms\n",
//	        result.Timestamp, result.Success, result.Duration/1000000)
//	}
func (c *Client) GetSuiteStatus(ctx context.Context, group, name string) (*SuiteStatus, error) {
	if name == "" {
		return nil, &ValidationError{
			Field:   "name",
			Message: "cannot be empty",
		}
	}
	key := GenerateKey(group, name)
	return c.GetSuiteStatusByKey(ctx, key)
}
