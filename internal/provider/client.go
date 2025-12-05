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

	// Read body once for logging (and error checking)
	var respBody []byte
	if resp.Body != nil {
		respBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		// Restore body for caller
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}

	// Log response with body
	c.logResponseWithBody(ctx, resp.StatusCode, resp.Status, respBody)

	return resp, nil
}

// checkResponse validates HTTP response status and returns detailed error
func (c *EmporixClient) checkResponse(ctx context.Context, statusCode int, body []byte, expectedStatuses ...int) error {
	for _, expected := range expectedStatuses {
		if statusCode == expected {
			return nil
		}
	}

	return fmt.Errorf("unexpected status code: %d, body: %s", statusCode, string(body))
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

func (c *EmporixClient) logResponseWithBody(ctx context.Context, statusCode int, status string, bodyBytes []byte) {
	if ctx == nil {
		return
	}

	// Log response metadata at DEBUG level
	tflog.Debug(ctx, "API Response",
		map[string]interface{}{
			"subsystem":   "http",
			"status":      status,
			"status_code": statusCode,
		})

	// Log response body at TRACE level
	if len(bodyBytes) > 0 {
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
}

func (c *EmporixClient) CreateSite(ctx context.Context, site *SiteSettings) error {
	path := fmt.Sprintf("/site/%s/sites", strings.ToLower(c.Tenant))
	resp, err := c.doRequest(ctx, "POST", path, site)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read body for error checking (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	return c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusCreated)
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

	// Read body (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var site SiteSettings
	if err := json.Unmarshal(bodyBytes, &site); err != nil {
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

	// Read body (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	return c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusNoContent, http.StatusOK)
}

func (c *EmporixClient) DeleteSite(ctx context.Context, siteCode string) error {
	path := fmt.Sprintf("/site/%s/sites/%s", strings.ToLower(c.Tenant), siteCode)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read body (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	return c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusNoContent, http.StatusOK)
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

	// Read body (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK, http.StatusNoContent); err != nil {
		return fmt.Errorf("failed to patch site mixins: %w", err)
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

	// Read body (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK, http.StatusNoContent); err != nil {
		return fmt.Errorf("failed to delete mixin %s: %w", mixinName, err)
	}

	return nil
}
