package gatussdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GetAllEndpointStatuses retrieves the status of all configured endpoints.
//
// Example:
//
//	statuses, err := client.GetAllEndpointStatuses(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, status := range statuses {
//	    fmt.Printf("Endpoint: %s (Key: %s)\n", status.Name, status.Key)
//	}
func (c *Client) GetAllEndpointStatuses(ctx context.Context) ([]EndpointStatus, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/endpoints/statuses")
	if err != nil {
		return nil, err
	}
	var statuses []EndpointStatus
	if err := c.decodeResponse(resp, &statuses); err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetEndpointStatusByKey retrieves the status of a specific endpoint by its key.
// The key should be in the format: {group}_{name}.
//
// Example:
//
//	status, err := client.GetEndpointStatusByKey(context.Background(), "core_blog-home")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Endpoint %s is healthy: %v\n", status.Name, status.Results[0].Success)
func (c *Client) GetEndpointStatusByKey(ctx context.Context, key string) (*EndpointStatus, error) {
	if key == "" {
		return nil, &ValidationError{
			Field:   "key",
			Message: "cannot be empty",
		}
	}
	path := fmt.Sprintf("/api/v1/endpoints/%s/statuses", url.PathEscape(key))
	resp, err := c.doRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	var status EndpointStatus
	if err := c.decodeResponse(resp, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// GetEndpointStatus retrieves the status of a specific endpoint by its group and name.
// The key is generated internally using GenerateKey.
//
// Example:
//
//	status, err := client.GetEndpointStatus(context.Background(), "core", "blog-home")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Endpoint %s is healthy: %v\n", status.Name, status.Results[0].Success)
func (c *Client) GetEndpointStatus(ctx context.Context, group, name string) (*EndpointStatus, error) {
	if name == "" {
		return nil, &ValidationError{
			Field:   "name",
			Message: "cannot be empty",
		}
	}
	key := GenerateKey(group, name)
	return c.GetEndpointStatusByKey(ctx, key)
}

// GetEndpointUptimeBadgeURL returns the URL for an endpoint's uptime badge.
// This method does not make an HTTP request, it just constructs the URL.
// Duration must be one of: 1h, 24h, 7d, 30d.
//
// Example:
//
//	url := client.GetEndpointUptimeBadgeURL("core_blog-home", "24h")
//	// Use the URL in markdown: ![Uptime](url)
func (c *Client) GetEndpointUptimeBadgeURL(key string, duration string) string {
	return fmt.Sprintf("%s/api/v1/endpoints/%s/uptimes/%s/badge.svg", c.baseURL, url.PathEscape(key), url.PathEscape(duration))
}

// GetEndpointHealthBadgeURL returns the URL for an endpoint's health badge.
// This method does not make an HTTP request, it just constructs the URL.
//
// Example:
//
//	url := client.GetEndpointHealthBadgeURL("core_blog-home")
//	// Use the URL in markdown: ![Health](url)
func (c *Client) GetEndpointHealthBadgeURL(key string) string {
	return fmt.Sprintf("%s/api/v1/endpoints/%s/health/badge.svg", c.baseURL, url.PathEscape(key))
}

// GetEndpointResponseTimeBadgeURL returns the URL for an endpoint's response time badge.
// This method does not make an HTTP request, it just constructs the URL.
// Duration must be one of: 1h, 24h, 7d, 30d.
//
// Example:
//
//	url := client.GetEndpointResponseTimeBadgeURL("core_blog-home", "24h")
//	// Use the URL in markdown: ![Response Time](url)
func (c *Client) GetEndpointResponseTimeBadgeURL(key string, duration string) string {
	return fmt.Sprintf("%s/api/v1/endpoints/%s/response-times/%s/badge.svg", c.baseURL, url.PathEscape(key), url.PathEscape(duration))
}

// GetEndpointUptime retrieves the uptime percentage for a specific endpoint.
// Duration must be one of: 1h, 24h, 7d, 30d.
//
// Example:
//
//	uptime, err := client.GetEndpointUptime(context.Background(), "core_blog-home", "24h")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Uptime: %.2f%%\n", uptime)
func (c *Client) GetEndpointUptime(ctx context.Context, key string, duration string) (float64, error) {
	uptimeData, err := c.GetEndpointUptimeData(ctx, key, duration)
	if err != nil {
		return 0, err
	}
	return uptimeData.Uptime, nil
}

// GetEndpointResponseTimes retrieves response time statistics for a specific endpoint.
// Duration must be one of: 1h, 24h, 7d, 30d.
//
// Example:
//
//	respTimes, err := client.GetEndpointResponseTimes(context.Background(), "core_blog-home", "24h")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Average: %dms, Min: %dms, Max: %dms\n",
//	    respTimes.Average/1000000, respTimes.Min/1000000, respTimes.Max/1000000)
func (c *Client) GetEndpointResponseTimes(ctx context.Context, key string, duration string) (*ResponseTimeData, error) {
	if key == "" {
		return nil, &ValidationError{
			Field:   "key",
			Message: "cannot be empty",
		}
	}
	path := fmt.Sprintf("/api/v1/endpoints/%s/response-times/%s", url.PathEscape(key), url.PathEscape(duration))
	resp, err := c.doRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	var data ResponseTimeData
	if err := c.decodeResponse(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetEndpointUptimeData retrieves raw uptime data for a specific endpoint.
// Duration must be one of: 1h, 24h, 7d, 30d.
//
// Example:
//
//	uptimeData, err := client.GetEndpointUptimeData(context.Background(), "core_blog-home", "24h")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Uptime: %.2f%% over %s\n", uptimeData.Uptime, uptimeData.Duration)
func (c *Client) GetEndpointUptimeData(ctx context.Context, key string, duration string) (*UptimeData, error) {
	if key == "" {
		return nil, &ValidationError{
			Field:   "key",
			Message: "cannot be empty",
		}
	}
	path := fmt.Sprintf("/api/v1/endpoints/%s/uptimes/%s", url.PathEscape(key), url.PathEscape(duration))
	resp, err := c.doRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	// Try to decode as UptimeData first
	var data UptimeData
	if err := c.decodeResponse(resp, &data); err != nil {
		// If that fails, try to decode as a simple float
		// (some Gatus versions return just the percentage)
		resp2, err2 := c.doRequest(ctx, http.MethodGet, path)
		if err2 != nil {
			return nil, err // Return original error
		}
		var uptimeFloat float64
		if err2 := c.decodeResponse(resp2, &uptimeFloat); err2 != nil {
			// If both fail, it might be an error response
			// Check if the original error was an API error
			var apiErr *APIError
			if errors.As(err, &apiErr) {
				return nil, apiErr
			}
			return nil, err // Return original error
		}
		// If we got a simple float, wrap it in UptimeData
		data = UptimeData{
			Uptime:   uptimeFloat,
			Duration: duration,
		}
	}
	return &data, nil
}

// PushExternalEndpointResult pushes a monitoring result to an external endpoint in Gatus.
// This is used for push-based monitoring where external systems can report their health status to Gatus.
// The endpoint must be configured as an external endpoint in Gatus with a matching token.
//
// Parameters:
//   - key: The endpoint key in the format {group}_{name} (use GenerateKey to create it)
//   - token: The bearer token configured for the external endpoint
//   - success: Whether the health check was successful
//   - errorMessage: Optional error message if the check failed (can be empty for successful checks)
//   - duration: Optional duration of the health check (e.g. "10s", "500ms")
//
// Example:
//
//	err := client.PushExternalEndpointResult(context.Background(), "core_ext-ep-test", "potato", true, "", "10s")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *Client) PushExternalEndpointResult(ctx context.Context, key string, token string, success bool, errorMessage string, duration string) error {
	if key == "" {
		return &ValidationError{
			Field:   "key",
			Message: "cannot be empty",
		}
	}
	if token == "" {
		return &ValidationError{
			Field:   "token",
			Message: "cannot be empty",
		}
	}
	// Build query parameters
	params := url.Values{}
	params.Set("success", fmt.Sprintf("%v", success))
	if errorMessage != "" {
		params.Set("error", errorMessage)
	}
	if duration != "" {
		params.Set("duration", duration)
	}
	path := fmt.Sprintf("/api/v1/endpoints/%s/external?%s", url.PathEscape(key), params.Encode())
	resp, err := c.doRequestWithAuth(ctx, http.MethodPost, path, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Check for success status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    http.StatusText(resp.StatusCode),
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}
	return nil
}
