package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type EmporixClient struct {
	Tenant      string
	AccessToken string
	ApiUrl      string
	Debug       bool // Enable detailed HTTP logging
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

	// Log request if debug is enabled or TF_LOG is set
	if c.Debug || os.Getenv("TF_LOG") == "TRACE" || os.Getenv("TF_LOG") == "DEBUG" {
		c.logRequest(req, bodyBytes)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	// Log response if debug is enabled
	if c.Debug || os.Getenv("TF_LOG") == "TRACE" || os.Getenv("TF_LOG") == "DEBUG" {
		c.logResponse(resp)
	}

	return resp, nil
}

func (c *EmporixClient) logRequest(req *http.Request, bodyBytes []byte) {
	fmt.Fprintf(os.Stderr, "\n[DEBUG] Emporix API Request:\n")
	fmt.Fprintf(os.Stderr, "================================================================================\n")
	fmt.Fprintf(os.Stderr, "%s %s\n", req.Method, req.URL.String())
	fmt.Fprintf(os.Stderr, "Headers:\n")
	for key, values := range req.Header {
		for _, value := range values {
			// Mask the authorization token
			if key == "Authorization" {
				fmt.Fprintf(os.Stderr, "  %s: Bearer ***REDACTED***\n", key)
			} else {
				fmt.Fprintf(os.Stderr, "  %s: %s\n", key, value)
			}
		}
	}
	if len(bodyBytes) > 0 {
		fmt.Fprintf(os.Stderr, "Body:\n")
		// Pretty print JSON
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyBytes, "  ", "  "); err == nil {
			fmt.Fprintf(os.Stderr, "%s\n", prettyJSON.String())
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", string(bodyBytes))
		}
	}
	fmt.Fprintf(os.Stderr, "================================================================================\n\n")
}

func (c *EmporixClient) logResponse(resp *http.Response) {
	fmt.Fprintf(os.Stderr, "\n[DEBUG] Emporix API Response:\n")
	fmt.Fprintf(os.Stderr, "================================================================================\n")
	fmt.Fprintf(os.Stderr, "Status: %s\n", resp.Status)
	fmt.Fprintf(os.Stderr, "Headers:\n")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", key, value)
		}
	}

	// Read and log body, then restore it for further processing
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil && len(bodyBytes) > 0 {
			fmt.Fprintf(os.Stderr, "Body:\n")
			// Pretty print JSON
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, bodyBytes, "  ", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", prettyJSON.String())
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", string(bodyBytes))
			}
			// Restore body for reading by caller
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
	fmt.Fprintf(os.Stderr, "================================================================================\n\n")
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
