package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NotFoundError is returned when a resource is not found (404)
type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "not found"
}

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}

// Global mutex map for per-tenant shipping zone operations
var (
	shippingZoneMutexes     = make(map[string]*sync.Mutex)
	shippingZoneMutexesLock sync.Mutex
)

// getShippingZoneMutex returns the mutex for a specific tenant's shipping zone operations
func getShippingZoneMutex(tenant string) *sync.Mutex {
	shippingZoneMutexesLock.Lock()
	defer shippingZoneMutexesLock.Unlock()

	if _, exists := shippingZoneMutexes[tenant]; !exists {
		shippingZoneMutexes[tenant] = &sync.Mutex{}
	}
	return shippingZoneMutexes[tenant]
}

// Global mutex map for per-tenant shipping method operations
var (
	shippingMethodMutexes     = make(map[string]*sync.Mutex)
	shippingMethodMutexesLock sync.Mutex
)

// getShippingMethodMutex returns the mutex for a specific tenant's shipping method operations
func getShippingMethodMutex(tenant string) *sync.Mutex {
	shippingMethodMutexesLock.Lock()
	defer shippingMethodMutexesLock.Unlock()

	if _, exists := shippingMethodMutexes[tenant]; !exists {
		shippingMethodMutexes[tenant] = &sync.Mutex{}
	}
	return shippingMethodMutexes[tenant]
}

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
		return nil, &NotFoundError{}
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
		return nil, &NotFoundError{}
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{}
	}

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

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{}
	}

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

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{}
	}

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

// CreateShippingZone creates a new shipping zone
func (c *EmporixClient) CreateShippingZone(ctx context.Context, site string, zone *ShippingZone) (*ShippingZone, error) {
	// Lock for this tenant's shipping zone operations
	mu := getShippingZoneMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones", strings.ToLower(c.Tenant), site)

	resp, err := c.doRequest(ctx, "POST", path, zone, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// Accept both 201 Created and 200 OK for successful creates
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusCreated, http.StatusOK); err != nil {
		return nil, err
	}

	// If response body is empty, return nil
	// The caller will do a GET to retrieve the actual state
	if len(bodyBytes) == 0 {
		return nil, nil
	}

	var createdZone ShippingZone
	if err := json.Unmarshal(bodyBytes, &createdZone); err != nil {
		// If unmarshal fails but create was successful, return nil
		// The caller will do a GET to retrieve the actual state
		tflog.Debug(ctx, "Failed to unmarshal create response, will rely on read-after-write", map[string]interface{}{
			"error": err.Error(),
			"body":  string(bodyBytes),
		})
		return nil, nil
	}

	return &createdZone, nil
}

// GetShippingZone retrieves a shipping zone by ID
func (c *EmporixClient) GetShippingZone(ctx context.Context, site, zoneID string) (*ShippingZone, error) {
	// Lock for this tenant's shipping zone operations
	mu := getShippingZoneMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones/%s", strings.ToLower(c.Tenant), site, zoneID)

	resp, err := c.doRequest(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var zone ShippingZone
	if err := json.Unmarshal(bodyBytes, &zone); err != nil {
		return nil, fmt.Errorf("error decoding shipping zone: %w", err)
	}

	return &zone, nil
}

// ListShippingZones retrieves all shipping zones for a site
func (c *EmporixClient) ListShippingZones(ctx context.Context, site string) ([]ShippingZone, error) {
	// Lock for this tenant's shipping zone operations
	mu := getShippingZoneMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones", strings.ToLower(c.Tenant), site)

	resp, err := c.doRequest(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var zones []ShippingZone
	if err := json.Unmarshal(bodyBytes, &zones); err != nil {
		return nil, fmt.Errorf("error decoding shipping zones list: %w", err)
	}

	return zones, nil
}

// UpdateShippingZone updates a shipping zone
func (c *EmporixClient) UpdateShippingZone(ctx context.Context, site, zoneID string, zone *ShippingZone) (*ShippingZone, error) {
	// Lock for this tenant's shipping zone operations
	mu := getShippingZoneMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones/%s", strings.ToLower(c.Tenant), site, zoneID)

	resp, err := c.doRequest(ctx, "PUT", path, zone, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// Accept both 200 OK and 204 No Content for successful updates
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK, http.StatusNoContent); err != nil {
		return nil, err
	}

	// If response is 204 No Content or empty body, return nil
	// The caller will do a GET to retrieve the actual state
	if resp.StatusCode == http.StatusNoContent || len(bodyBytes) == 0 {
		return nil, nil
	}

	var updatedZone ShippingZone
	if err := json.Unmarshal(bodyBytes, &updatedZone); err != nil {
		// If unmarshal fails but update was successful, return nil
		// The caller will do a GET to retrieve the actual state
		tflog.Debug(ctx, "Failed to unmarshal update response, will rely on read-after-write", map[string]interface{}{
			"error": err.Error(),
			"body":  string(bodyBytes),
		})
		return nil, nil
	}

	return &updatedZone, nil
}

// DeleteShippingZone deletes a shipping zone
func (c *EmporixClient) DeleteShippingZone(ctx context.Context, site, zoneID string) error {
	// Lock for this tenant's shipping zone operations
	mu := getShippingZoneMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones/%s", strings.ToLower(c.Tenant), site, zoneID)

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

// CreateSchema creates a new schema
func (c *EmporixClient) CreateSchema(ctx context.Context, schema *SchemaCreate) (*Schema, error) {
	path := fmt.Sprintf("/schema/%s/schemas", strings.ToLower(c.Tenant))

	// Name is always a map, so always use Content-Language: *
	headers := map[string]string{
		"Content-Language": "*",
	}

	resp, err := c.doRequest(ctx, "POST", path, schema, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusCreated); err != nil {
		return nil, err
	}

	// API returns IdResponse on creation
	var idResp IdResponse
	if err := json.Unmarshal(bodyBytes, &idResp); err != nil {
		return nil, fmt.Errorf("error decoding schema creation response: %w", err)
	}

	tflog.Debug(ctx, "Schema created, fetching complete state via GET")

	// Fetch complete schema data
	return c.GetSchema(ctx, idResp.ID)
}

// GetSchema retrieves a schema by ID
func (c *EmporixClient) GetSchema(ctx context.Context, id string) (*Schema, error) {
	path := fmt.Sprintf("/schema/%s/schemas/%s", strings.ToLower(c.Tenant), id)

	// Always use Accept-Language: * to retrieve all translations
	headers := map[string]string{
		"Accept-Language": "*",
	}

	resp, err := c.doRequest(ctx, "GET", path, nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var schema Schema
	if err := json.Unmarshal(bodyBytes, &schema); err != nil {
		return nil, fmt.Errorf("error decoding schema: %w", err)
	}

	return &schema, nil
}

// UpdateSchema updates a schema
func (c *EmporixClient) UpdateSchema(ctx context.Context, id string, updateData *SchemaUpdate) (*Schema, error) {
	// First, get current schema to retrieve metadata.version (required for PUT)
	schema, err := c.GetSchema(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting schema before update: %w", err)
	}

	// Add metadata.version to update data (required by API)
	if schema.Metadata != nil && schema.Metadata.Version > 0 {
		if updateData.Metadata == nil {
			updateData.Metadata = &SchemaMetadataUpdate{}
		}
		updateData.Metadata.Version = schema.Metadata.Version
	}

	path := fmt.Sprintf("/schema/%s/schemas/%s", strings.ToLower(c.Tenant), id)

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

	return c.GetSchema(ctx, id)
}

// DeleteSchema deletes a schema
func (c *EmporixClient) DeleteSchema(ctx context.Context, id string) error {
	path := fmt.Sprintf("/schema/%s/schemas/%s", strings.ToLower(c.Tenant), id)

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

// DeliveryTime represents a delivery time configuration
// CreateDeliveryTime creates a new delivery time
func (c *EmporixClient) CreateDeliveryTime(ctx context.Context, deliveryTime *DeliveryTime) (*DeliveryTime, error) {
	path := fmt.Sprintf("/shipping/%s/delivery-times", strings.ToLower(c.Tenant))

	resp, err := c.doRequest(ctx, "POST", path, deliveryTime, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// Accept both 201 Created and 200 OK
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusCreated, http.StatusOK); err != nil {
		return nil, err
	}

	// If response body is empty, return nil
	if len(bodyBytes) == 0 {
		return nil, nil
	}

	var createdDeliveryTime DeliveryTime
	if err := json.Unmarshal(bodyBytes, &createdDeliveryTime); err != nil {
		tflog.Debug(ctx, "Failed to unmarshal create response, will rely on read-after-write", map[string]interface{}{
			"error": err.Error(),
			"body":  string(bodyBytes),
		})
		return nil, nil
	}

	return &createdDeliveryTime, nil
}

// GetDeliveryTime retrieves a delivery time by ID
func (c *EmporixClient) GetDeliveryTime(ctx context.Context, id string) (*DeliveryTime, error) {
	path := fmt.Sprintf("/shipping/%s/delivery-times/%s", strings.ToLower(c.Tenant), id)

	resp, err := c.doRequest(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var deliveryTime DeliveryTime
	if err := json.Unmarshal(bodyBytes, &deliveryTime); err != nil {
		return nil, fmt.Errorf("error decoding delivery time: %w", err)
	}

	return &deliveryTime, nil
}

// UpdateDeliveryTime updates a delivery time
func (c *EmporixClient) UpdateDeliveryTime(ctx context.Context, id string, deliveryTime *DeliveryTime) (*DeliveryTime, error) {
	path := fmt.Sprintf("/shipping/%s/delivery-times/%s", strings.ToLower(c.Tenant), id)

	resp, err := c.doRequest(ctx, "PUT", path, deliveryTime, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// Accept both 200 OK and 204 No Content
	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK, http.StatusNoContent); err != nil {
		return nil, err
	}

	// If response is 204 No Content or empty body, return nil
	if resp.StatusCode == http.StatusNoContent || len(bodyBytes) == 0 {
		return nil, nil
	}

	var updatedDeliveryTime DeliveryTime
	if err := json.Unmarshal(bodyBytes, &updatedDeliveryTime); err != nil {
		tflog.Debug(ctx, "Failed to unmarshal update response, will rely on read-after-write", map[string]interface{}{
			"error": err.Error(),
			"body":  string(bodyBytes),
		})
		return nil, nil
	}

	return &updatedDeliveryTime, nil
}

// DeleteDeliveryTime deletes a delivery time
func (c *EmporixClient) DeleteDeliveryTime(ctx context.Context, id string) error {
	path := fmt.Sprintf("/shipping/%s/delivery-times/%s", strings.ToLower(c.Tenant), id)

	resp, err := c.doRequest(ctx, "DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusNoContent, http.StatusOK); err != nil {
		return err
	}

	return nil
}

// ShippingMethod represents a shipping method
// CreateShippingMethod creates a new shipping method
func (c *EmporixClient) CreateShippingMethod(ctx context.Context, site, zoneID string, shippingMethod *ShippingMethod) (*ShippingMethod, error) {
	// Lock for this tenant's shipping method operations
	mu := getShippingMethodMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones/%s/methods", strings.ToLower(c.Tenant), site, zoneID)

	resp, err := c.doRequest(ctx, "POST", path, shippingMethod, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusCreated, http.StatusOK); err != nil {
		return nil, err
	}

	var created ShippingMethod
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("error decoding shipping method: %w", err)
	}

	return &created, nil
}

// GetShippingMethod retrieves a shipping method by ID
func (c *EmporixClient) GetShippingMethod(ctx context.Context, site, zoneID, id string) (*ShippingMethod, error) {
	// Lock for this tenant's shipping method operations
	mu := getShippingMethodMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones/%s/methods/%s", strings.ToLower(c.Tenant), site, zoneID, id)

	resp, err := c.doRequest(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK); err != nil {
		return nil, err
	}

	var shippingMethod ShippingMethod
	if err := json.Unmarshal(bodyBytes, &shippingMethod); err != nil {
		return nil, fmt.Errorf("error decoding shipping method: %w", err)
	}

	return &shippingMethod, nil
}

// UpdateShippingMethod updates a shipping method
func (c *EmporixClient) UpdateShippingMethod(ctx context.Context, site, zoneID, id string, shippingMethod *ShippingMethod) (*ShippingMethod, error) {
	// Lock for this tenant's shipping method operations
	mu := getShippingMethodMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones/%s/methods/%s", strings.ToLower(c.Tenant), site, zoneID, id)

	resp, err := c.doRequest(ctx, "PUT", path, shippingMethod, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusOK, http.StatusNoContent); err != nil {
		return nil, err
	}

	// If response is 204 No Content or empty body, return nil
	if resp.StatusCode == http.StatusNoContent || len(bodyBytes) == 0 {
		return nil, nil
	}

	var updated ShippingMethod
	if err := json.Unmarshal(bodyBytes, &updated); err != nil {
		tflog.Debug(ctx, "Failed to unmarshal update response, will rely on read-after-write", map[string]interface{}{
			"error": err.Error(),
			"body":  string(bodyBytes),
		})
		return nil, nil
	}

	return &updated, nil
}

// DeleteShippingMethod deletes a shipping method
func (c *EmporixClient) DeleteShippingMethod(ctx context.Context, site, zoneID, id string) error {
	// Lock for this tenant's shipping method operations
	mu := getShippingMethodMutex(c.Tenant)
	mu.Lock()
	defer mu.Unlock()

	path := fmt.Sprintf("/shipping/%s/%s/zones/%s/methods/%s", strings.ToLower(c.Tenant), site, zoneID, id)

	resp, err := c.doRequest(ctx, "DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := c.checkResponse(ctx, resp.StatusCode, bodyBytes, http.StatusNoContent, http.StatusOK); err != nil {
		return err
	}

	return nil
}
