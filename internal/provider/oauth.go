package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func generateAccessToken(ctx context.Context, apiUrl, clientId, clientSecret, scope string) (string, error) {
	tokenURL := apiUrl + "/oauth/token"

	// Prepare form data
	formData := url.Values{}
	formData.Set("client_id", clientId)
	formData.Set("client_secret", clientSecret)
	formData.Set("grant_type", "client_credentials")

	// Only add scope if provided
	if scope != "" {
		formData.Set("scope", scope)
	}

	logFields := map[string]interface{}{
		"subsystem":  "oauth",
		"url":        tokenURL,
		"grant_type": "client_credentials",
	}
	if scope != "" {
		logFields["scope"] = scope
	}

	tflog.Debug(ctx, "Requesting OAuth access token", logFields)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making token request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		tflog.Error(ctx, "OAuth token request failed",
			map[string]interface{}{
				"subsystem":   "oauth",
				"status_code": resp.StatusCode,
				"status":      resp.Status,
				"response":    string(respBody),
			})
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResponse OAuthTokenResponse
	if err := json.Unmarshal(respBody, &tokenResponse); err != nil {
		return "", fmt.Errorf("error parsing token response: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	tflog.Debug(ctx, "Successfully obtained OAuth access token",
		map[string]interface{}{
			"subsystem":  "oauth",
			"expires_in": tokenResponse.ExpiresIn,
			"token_type": tokenResponse.TokenType,
		})

	return tokenResponse.AccessToken, nil
}
