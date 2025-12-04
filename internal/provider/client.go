package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type EmporixClient struct {
	Tenant      string
	AccessToken string
	ApiUrl      string
	httpClient  *http.Client
}

func NewEmporixClient(tenant, accessToken, apiUrl string) *EmporixClient {
	return &EmporixClient{
		Tenant:      tenant,
		AccessToken: accessToken,
		ApiUrl:      apiUrl,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *EmporixClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
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

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	// Log request
	c.logRequest(ctx, req, bodyBytes)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	// Always log response (Terraform's TF_LOG level controls what's shown)
	c.logResponse(ctx, resp)

	return resp, nil
}

// checkResponse is a helper to validate HTTP response and return detailed error
func (c *EmporixClient) checkResponse(ctx context.Context, resp *http.Response, expectedStatus int) error {
	if resp.StatusCode == expectedStatus {
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	c.logResponseBody(ctx, resp.StatusCode, bodyBytes)

	return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
}

func (c *EmporixClient) logRequest(ctx context.Context, req *http.Request, bodyBytes []byte) {
	if ctx == nil {
		return
	}

	// Extract tf_req_id from context if available
	logFields := map[string]interface{}{
		"subsystem": "http",
		"method":    req.Method,
		"url":       req.URL.String(),
	}

	// Log with http subsystem
	tflog.Debug(ctx, "API Request", logFields)

	// Log body if present - pretty print for readability
	if len(bodyBytes) > 0 {
		// Pretty print JSON
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
			tflog.Trace(ctx, "Request body (JSON):\n"+prettyJSON.String(), map[string]interface{}{"subsystem": "http"})
		} else {
			tflog.Trace(ctx, "Request body", map[string]interface{}{
				"subsystem": "http",
				"body":      string(bodyBytes),
			})
		}
	}
}

func (c *EmporixClient) logResponse(ctx context.Context, resp *http.Response) {
	if ctx == nil || resp == nil {
		return
	}

	// Log with http subsystem
	tflog.Debug(ctx, "API Response",
		map[string]interface{}{
			"subsystem":   "http",
			"status":      resp.Status,
			"status_code": resp.StatusCode,
		})

	// Read and log body, then restore it
	if resp.Body != nil {
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil && len(bodyBytes) > 0 {
			c.logResponseBody(ctx, resp.StatusCode, bodyBytes)
			// Restore body
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
}

func (c *EmporixClient) logResponseBody(ctx context.Context, statusCode int, bodyBytes []byte) {
	if len(bodyBytes) == 0 {
		return
	}

	// Pretty print JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
		tflog.Trace(ctx, "Response body (JSON):\n"+prettyJSON.String(), map[string]interface{}{
			"subsystem":   "http",
			"status_code": statusCode,
		})
	} else {
		tflog.Trace(ctx, "Response body", map[string]interface{}{
			"subsystem":   "http",
			"status_code": statusCode,
			"body":        string(bodyBytes),
		})
	}
}

func (c *EmporixClient) CreateSite(ctx context.Context, site *SiteSettings) error {
	path := fmt.Sprintf("/site/%s/sites", strings.ToLower(c.Tenant))
	resp, err := c.doRequest(ctx, "POST", path, site)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return c.checkResponse(ctx, resp, http.StatusCreated)
}

func (c *EmporixClient) GetSite(ctx context.Context, siteCode string) (*SiteSettings, error) {
	path := fmt.Sprintf("/site/%s/sites/%s?expand=mixin:*", strings.ToLower(c.Tenant), siteCode)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if err := c.checkResponse(ctx, resp, http.StatusOK); err != nil {
		return nil, err
	}

	var site SiteSettings
	if err := json.NewDecoder(resp.Body).Decode(&site); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &site, nil
}

func (c *EmporixClient) UpdateSite(ctx context.Context, siteCode string, patchData map[string]interface{}) error {
	path := fmt.Sprintf("/site/%s/sites/%s", strings.ToLower(c.Tenant), siteCode)
	resp, err := c.doRequest(ctx, "PATCH", path, patchData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logResponseBody(ctx, resp.StatusCode, bodyBytes)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *EmporixClient) DeleteSite(ctx context.Context, siteCode string) error {
	path := fmt.Sprintf("/site/%s/sites/%s", strings.ToLower(c.Tenant), siteCode)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logResponseBody(ctx, resp.StatusCode, bodyBytes)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// PatchSiteMixins updates mixins and metadata using PATCH
func (c *EmporixClient) PatchSiteMixins(ctx context.Context, siteCode string, mixins map[string]interface{}, metadata *Metadata) error {
	path := fmt.Sprintf("/site/%s/sites/%s", strings.ToLower(c.Tenant), siteCode)

	patchData := make(map[string]interface{}, 2)

	if mixins != nil {
		patchData["mixins"] = mixins
	}

	// Only include metadata.mixins (schema URLs), NOT version
	if metadata != nil && metadata.Mixins != nil && len(metadata.Mixins) > 0 {
		patchData["metadata"] = map[string]interface{}{
			"mixins": metadata.Mixins,
		}
	}

	resp, err := c.doRequest(ctx, "PATCH", path, patchData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logResponseBody(ctx, resp.StatusCode, bodyBytes)
		return fmt.Errorf("failed to patch site mixins: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteSiteMixin deletes a specific mixin by name
func (c *EmporixClient) DeleteSiteMixin(ctx context.Context, siteCode string, mixinName string) error {
	path := fmt.Sprintf("/site/%s/sites/%s/mixins/%s", strings.ToLower(c.Tenant), siteCode, mixinName)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logResponseBody(ctx, resp.StatusCode, bodyBytes)
		return fmt.Errorf("failed to delete mixin %s: status %d, body: %s", mixinName, resp.StatusCode, string(bodyBytes))
	}

	return nil
}
