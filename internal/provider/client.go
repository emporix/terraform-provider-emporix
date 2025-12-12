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

func (c *EmporixClient) doRequest(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
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

	// Add custom headers if provided
	for key, value := range headers {
		req.Header.Set(key, value)
	}

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
	resp, err := c.doRequest(ctx, "POST", path, site, nil)
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
	resp, err := c.doRequest(ctx, "GET", path, nil, nil)
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
	resp, err := c.doRequest(ctx, "PATCH", path, patchData, nil)
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
	resp, err := c.doRequest(ctx, "DELETE", path, nil, nil)
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

	resp, err := c.doRequest(ctx, "PATCH", path, patchData, nil)
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
	resp, err := c.doRequest(ctx, "DELETE", path, nil, nil)
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

// CreatePaymentMode creates a new payment mode
func (c *EmporixClient) CreatePaymentMode(ctx context.Context, paymentMode *PaymentMode) (*PaymentMode, error) {
	path := fmt.Sprintf("/payment-gateway/%s/paymentmodes/config", strings.ToLower(c.Tenant))
	resp, err := c.doRequest(ctx, "POST", path, paymentMode, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read body (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var created PaymentMode
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &created, nil
}

// GetPaymentMode retrieves a single payment mode by ID
func (c *EmporixClient) GetPaymentMode(ctx context.Context, id string) (*PaymentMode, error) {
	path := fmt.Sprintf("/payment-gateway/%s/paymentmodes/config/%s", strings.ToLower(c.Tenant), id)
	resp, err := c.doRequest(ctx, "GET", path, nil, nil)
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

	var paymentMode PaymentMode
	if err := json.Unmarshal(bodyBytes, &paymentMode); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &paymentMode, nil
}

// UpdatePaymentMode updates an existing payment mode
func (c *EmporixClient) UpdatePaymentMode(ctx context.Context, id string, updateData *PaymentModeUpdate) (*PaymentMode, error) {
	path := fmt.Sprintf("/payment-gateway/%s/paymentmodes/config/%s", strings.ToLower(c.Tenant), id)
	resp, err := c.doRequest(ctx, "PUT", path, updateData, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read body (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	// WORKAROUND: The API returns stale data in the PUT response.
	// Do a separate GET request to fetch the actual updated state.
	// This prevents "Provider produced inconsistent result after apply" errors.
	tflog.Debug(ctx, "Update succeeded, fetching current state via GET (API returns stale data in PUT response)")

	return c.GetPaymentMode(ctx, id)
}

// DeletePaymentMode deletes a payment mode
func (c *EmporixClient) DeletePaymentMode(ctx context.Context, id string) error {
	path := fmt.Sprintf("/payment-gateway/%s/paymentmodes/config/%s", strings.ToLower(c.Tenant), id)
	resp, err := c.doRequest(ctx, "DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read body (already logged in doRequest)
	bodyBytes, _ := io.ReadAll(resp.Body)
	return c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK, http.StatusNoContent)
}

// GetCountry retrieves a country by code
func (c *EmporixClient) GetCountry(ctx context.Context, code string) (*Country, error) {
	path := fmt.Sprintf("/country/%s/countries/%s", strings.ToLower(c.Tenant), code)

	headers := map[string]string{
		"X-Version": "v2",
	}

	resp, err := c.doRequest(ctx, "GET", path, nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var country Country
	if err := json.Unmarshal(bodyBytes, &country); err != nil {
		return nil, fmt.Errorf("error decoding country response: %w", err)
	}

	return &country, nil
}

// UpdateCountry updates a country's active status
func (c *EmporixClient) UpdateCountry(ctx context.Context, code string, updateData *CountryUpdate) (*Country, error) {
	// First, get current country to retrieve metadata.version (required for PATCH)
	country, err := c.GetCountry(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("error getting country before update: %w", err)
	}

	// Add metadata.version to update data (required by API)
	if country.Metadata != nil && country.Metadata.Version > 0 {
		if updateData.Metadata == nil {
			updateData.Metadata = &Metadata{}
		}
		updateData.Metadata.Version = country.Metadata.Version
	}

	path := fmt.Sprintf("/country/%s/countries/%s", strings.ToLower(c.Tenant), code)

	headers := map[string]string{
		"X-Version": "v2",
	}

	resp, err := c.doRequest(ctx, "PATCH", path, updateData, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusNoContent); err != nil {
		return nil, err
	}

	// PATCH returns 204 No Content, so fetch current state via GET
	tflog.Debug(ctx, "Update succeeded, fetching current state via GET")

	return c.GetCountry(ctx, code)
}

// CreateCurrency creates a new currency
func (c *EmporixClient) CreateCurrency(ctx context.Context, currency *CurrencyCreate) (*Currency, error) {
	path := fmt.Sprintf("/currency/%s/currencies", strings.ToLower(c.Tenant))

	// Name is always a map, so always use Content-Language: *
	headers := map[string]string{
		"Content-Language": "*",
	}

	resp, err := c.doRequest(ctx, "POST", path, currency, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusCreated); err != nil {
		return nil, err
	}

	// API may not return complete data in CREATE response (especially when using Content-Language: *)
	// Always fetch the currency after creation to get complete state with all translations
	var createdCurrency Currency
	if err := json.Unmarshal(bodyBytes, &createdCurrency); err != nil {
		return nil, fmt.Errorf("error decoding currency response: %w", err)
	}

	tflog.Debug(ctx, "Currency created, fetching complete state via GET")

	// Fetch complete currency data with all translations
	return c.GetCurrency(ctx, createdCurrency.Code)
}

// GetCurrency retrieves a currency by code
func (c *EmporixClient) GetCurrency(ctx context.Context, code string) (*Currency, error) {
	path := fmt.Sprintf("/currency/%s/currencies/%s", strings.ToLower(c.Tenant), code)

	// Always use Accept-Language: * to retrieve all translations
	headers := map[string]string{
		"Accept-Language": "*",
	}

	resp, err := c.doRequest(ctx, "GET", path, nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var currency Currency
	if err := json.Unmarshal(bodyBytes, &currency); err != nil {
		return nil, fmt.Errorf("error decoding currency response: %w", err)
	}

	return &currency, nil
}

// UpdateCurrency updates a currency
func (c *EmporixClient) UpdateCurrency(ctx context.Context, code string, updateData *CurrencyUpdate) (*Currency, error) {
	// First, get current currency to retrieve metadata.version (required for PUT)
	currency, err := c.GetCurrency(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("error getting currency before update: %w", err)
	}

	// Add metadata.version to update data (required by API)
	if currency.Metadata != nil && currency.Metadata.Version > 0 {
		if updateData.Metadata == nil {
			updateData.Metadata = &Metadata{}
		}
		updateData.Metadata.Version = currency.Metadata.Version
	}

	path := fmt.Sprintf("/currency/%s/currencies/%s", strings.ToLower(c.Tenant), code)

	// Name is always a map, so always use Content-Language: *
	headers := map[string]string{
		"Content-Language": "*",
	}

	resp, err := c.doRequest(ctx, "PUT", path, updateData, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusNoContent); err != nil {
		return nil, err
	}

	// PUT returns 204 No Content, so fetch current state via GET
	tflog.Debug(ctx, "Update succeeded, fetching current state via GET")

	return c.GetCurrency(ctx, code)
}

// DeleteCurrency deletes a currency by code
func (c *EmporixClient) DeleteCurrency(ctx context.Context, code string) error {
	path := fmt.Sprintf("/currency/%s/currencies/%s", strings.ToLower(c.Tenant), code)

	resp, err := c.doRequest(ctx, "DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusNoContent); err != nil {
		return err
	}

	return nil
}

// CreateTenantConfiguration creates a new tenant configuration
func (c *EmporixClient) CreateTenantConfiguration(ctx context.Context, config *TenantConfigurationCreate) (*TenantConfiguration, error) {
	path := fmt.Sprintf("/configuration/%s/configurations", strings.ToLower(c.Tenant))

	// Wrap in array since API expects array
	configs := []TenantConfigurationCreate{*config}

	resp, err := c.doRequest(ctx, "POST", path, configs, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusCreated); err != nil {
		return nil, err
	}

	// API returns array, we need the first element
	var createdConfigs []TenantConfiguration
	if err := json.Unmarshal(bodyBytes, &createdConfigs); err != nil {
		return nil, fmt.Errorf("error decoding tenant configuration response: %w", err)
	}

	if len(createdConfigs) == 0 {
		return nil, fmt.Errorf("API returned empty array")
	}

	return &createdConfigs[0], nil
}

// GetTenantConfiguration retrieves a tenant configuration by key
func (c *EmporixClient) GetTenantConfiguration(ctx context.Context, key string) (*TenantConfiguration, error) {
	path := fmt.Sprintf("/configuration/%s/configurations/%s", strings.ToLower(c.Tenant), key)

	resp, err := c.doRequest(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var config TenantConfiguration
	if err := json.Unmarshal(bodyBytes, &config); err != nil {
		return nil, fmt.Errorf("error decoding tenant configuration: %w", err)
	}

	return &config, nil
}

// UpdateTenantConfiguration updates a tenant configuration
func (c *EmporixClient) UpdateTenantConfiguration(ctx context.Context, key string, updateData *TenantConfigurationUpdate) (*TenantConfiguration, error) {
	path := fmt.Sprintf("/configuration/%s/configurations/%s", strings.ToLower(c.Tenant), key)

	resp, err := c.doRequest(ctx, "PUT", path, updateData, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var config TenantConfiguration
	if err := json.Unmarshal(bodyBytes, &config); err != nil {
		return nil, fmt.Errorf("error decoding tenant configuration: %w", err)
	}

	return &config, nil
}

// DeleteTenantConfiguration deletes a tenant configuration by key
func (c *EmporixClient) DeleteTenantConfiguration(ctx context.Context, key string) error {
	path := fmt.Sprintf("/configuration/%s/configurations/%s", strings.ToLower(c.Tenant), key)

	resp, err := c.doRequest(ctx, "DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusNoContent); err != nil {
		return err
	}

	return nil
}
