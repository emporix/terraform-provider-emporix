package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type EmporixClient struct {
	Tenant      string
	AccessToken string
	ApiUrl      string
	ctx         context.Context
}

func (c *EmporixClient) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *EmporixClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.ApiUrl, path)

	var bodyReader io.Reader
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	// Log request
	c.logRequest(req, bodyBytes)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	// Log response
	c.logResponse(resp)

	return resp, nil
}

func (c *EmporixClient) logRequest(req *http.Request, bodyBytes []byte) {
	if c.ctx == nil {
		return
	}

	// Extract tf_req_id from context if available
	logFields := map[string]interface{}{
		"subsystem": "http",
		"method":    req.Method,
		"url":       req.URL.String(),
	}

	// Log with http subsystem
	tflog.Debug(c.ctx, "API Request", logFields)

	// Log body if present - pretty print for readability
	if len(bodyBytes) > 0 {
		// Pretty print JSON
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
			tflog.Trace(c.ctx, "Request body (JSON):\n"+prettyJSON.String(), map[string]interface{}{"subsystem": "http"})
		} else {
			tflog.Trace(c.ctx, "Request body", map[string]interface{}{
				"subsystem": "http",
				"body":      string(bodyBytes),
			})
		}
	}
}

func (c *EmporixClient) logResponse(resp *http.Response) {
	if c.ctx == nil {
		return
	}

	// Log with http subsystem
	tflog.Debug(c.ctx, "API Response",
		map[string]interface{}{
			"subsystem":   "http",
			"status":      resp.Status,
			"status_code": resp.StatusCode,
		})

	// Read and log body, then restore it
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil && len(bodyBytes) > 0 {
			// Pretty print JSON
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
				tflog.Trace(c.ctx, "Response body (JSON):\n"+prettyJSON.String(), map[string]interface{}{"subsystem": "http"})
			} else {
				tflog.Trace(c.ctx, "Response body", map[string]interface{}{
					"subsystem": "http",
					"body":      string(bodyBytes),
				})
			}
			// Restore body
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
}

func (c *EmporixClient) CreateSite(site *SiteSettings) error {
	path := fmt.Sprintf("/site/%s/sites", strings.ToLower(c.Tenant))
	resp, err := c.doRequest("POST", path, site)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *EmporixClient) GetSite(siteCode string) (*SiteSettings, error) {
	path := fmt.Sprintf("/site/%s/sites/%s?expand=mixin:*", strings.ToLower(c.Tenant), siteCode)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var site SiteSettings
	if err := json.NewDecoder(resp.Body).Decode(&site); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &site, nil
}

func (c *EmporixClient) UpdateSite(siteCode string, site *SiteSettings) error {
	path := fmt.Sprintf("/site/%s/sites/%s", strings.ToLower(c.Tenant), siteCode)
	resp, err := c.doRequest("PUT", path, site)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *EmporixClient) DeleteSite(siteCode string) error {
	path := fmt.Sprintf("/site/%s/sites/%s", strings.ToLower(c.Tenant), siteCode)
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
