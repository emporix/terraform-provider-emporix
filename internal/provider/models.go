package provider

type SiteSettings struct {
	Code                      string                 `json:"code,omitempty"`
	Name                      string                 `json:"name"`
	Active                    bool                   `json:"active"`
	Default                   bool                   `json:"default"`
	IncludesTax               *bool                  `json:"includesTax,omitempty"`
	DefaultLanguage           string                 `json:"defaultLanguage"`
	Languages                 []string               `json:"languages"`
	Currency                  string                 `json:"currency"`
	AvailableCurrencies       []string               `json:"availableCurrencies,omitempty"`
	ShipToCountries           []string               `json:"shipToCountries,omitempty"`
	TaxCalculationAddressType string                 `json:"taxCalculationAddressType,omitempty"`
	DecimalPoints             *int64                 `json:"decimalPoints,omitempty"`
	CartCalculationScale      *int64                 `json:"cartCalculationScale,omitempty"`
	HomeBase                  *HomeBase              `json:"homeBase,omitempty"`
	AssistedBuying            *AssistedBuying        `json:"assistedBuying,omitempty"`
	Mixins                    map[string]interface{} `json:"mixins,omitempty"`
	Metadata                  *Metadata              `json:"metadata,omitempty"`
}

type HomeBase struct {
	Address  *Address  `json:"address,omitempty"`
	Location *Location `json:"location,omitempty"`
}

type Address struct {
	Street       string `json:"street,omitempty"`
	StreetNumber string `json:"streetNumber,omitempty"`
	ZipCode      string `json:"zipCode,omitempty"`
	City         string `json:"city,omitempty"`
	Country      string `json:"country"`
	State        string `json:"state,omitempty"`
}

type Location struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type AssistedBuying struct {
	StorefrontUrl string `json:"storefrontUrl,omitempty"`
}

type Metadata struct {
	Mixins  map[string]string `json:"mixins,omitempty"`
	Version int               `json:"version,omitempty"`
}

// PaymentMode represents a payment mode configuration
type PaymentMode struct {
	ID            string            `json:"id,omitempty"`
	Code          string            `json:"code"`
	Active        bool              `json:"active"`
	Provider      string            `json:"provider"`
	Configuration map[string]string `json:"configuration,omitempty"`
}

// PaymentModeUpdate represents the update payload for a payment mode
type PaymentModeUpdate struct {
	Active        bool              `json:"active"`
	Configuration map[string]string `json:"configuration,omitempty"`
}

// Country represents a country in Emporix
type Country struct {
	Code     string            `json:"code"`
	Name     map[string]string `json:"name"`
	Regions  []string          `json:"regions,omitempty"`
	Active   bool              `json:"active"`
	Metadata *Metadata         `json:"metadata,omitempty"`
}

// CountryUpdate represents data for updating a country (only active field can be updated)
type CountryUpdate struct {
	Active   *bool     `json:"active,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// Currency represents a currency in Emporix
type Currency struct {
	Code     string            `json:"code"`
	Name     map[string]string `json:"name"` // Always returned as map from API
	Metadata *Metadata         `json:"metadata,omitempty"`
}

// CurrencyCreate represents the creation payload for a currency
// Name can be either:
// - string (with Content-Language header set to specific language)
// - map[string]string (with Content-Language header set to *)
type CurrencyCreate struct {
	Code string      `json:"code"`
	Name interface{} `json:"name"` // string or map[string]string
}

// CurrencyUpdate represents the update payload for a currency
// Name can be either:
// - string (with Content-Language header set to specific language)
// - map[string]string (with Content-Language header set to *)
type CurrencyUpdate struct {
	Name     interface{} `json:"name"` // string or map[string]string
	Metadata *Metadata   `json:"metadata,omitempty"`
}

// TenantConfiguration represents a tenant configuration
type TenantConfiguration struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value"` // Can be object, string, array, or boolean
	Version int         `json:"version"`
	Secured bool        `json:"secured"`
}

// TenantConfigurationCreate represents the creation payload for a tenant configuration
type TenantConfigurationCreate struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
	Secured bool        `json:"secured,omitempty"`
}

// TenantConfigurationUpdate represents the update payload for a tenant configuration
type TenantConfigurationUpdate struct {
	Key     string      `json:"key,omitempty"`
	Value   interface{} `json:"value,omitempty"`
	Version int         `json:"version,omitempty"`
	Secured bool        `json:"secured,omitempty"`
}

// Tax represents a tax configuration in Emporix
type Tax struct {
	LocationCode string       `json:"locationCode"`
	Location     *TaxLocation `json:"location"`
	TaxClasses   []TaxClass   `json:"taxClasses"`
	Metadata     *Metadata    `json:"metadata,omitempty"`
}

// TaxLocation represents the location for a tax configuration
type TaxLocation struct {
	CountryCode string `json:"countryCode"`
}

// TaxClass represents a tax class within a tax configuration
type TaxClass struct {
	Code        string      `json:"code"`
	Name        interface{} `json:"name"` // string or map[string]string
	Rate        float64     `json:"rate"`
	Description interface{} `json:"description,omitempty"` // string or map[string]string
	Order       *int        `json:"order,omitempty"`
	IsDefault   bool        `json:"isDefault,omitempty"`
}

// TaxCreate represents the creation payload for a tax configuration
type TaxCreate struct {
	Location   *TaxLocation `json:"location"`
	TaxClasses []TaxClass   `json:"taxClasses"`
}

// TaxUpdate represents the update payload for a tax configuration
type TaxUpdate struct {
	Location   *TaxLocation `json:"location"`
	TaxClasses []TaxClass   `json:"taxClasses"`
	Metadata   *Metadata    `json:"metadata,omitempty"`
}

// Schema represents a schema in Emporix
type Schema struct {
	ID         string            `json:"id"`
	Name       map[string]string `json:"name"`
	Types      []string          `json:"types"`
	Attributes []SchemaAttribute `json:"attributes"`
	Metadata   *SchemaMetadata   `json:"metadata,omitempty"`
}

// SchemaCreate represents the creation payload for a schema
type SchemaCreate struct {
	ID         string            `json:"id,omitempty"`
	Name       map[string]string `json:"name"`
	Types      []string          `json:"types"`
	Attributes []SchemaAttribute `json:"attributes"`
}

// SchemaUpdate represents the update payload for a schema
type SchemaUpdate struct {
	Name       map[string]string     `json:"name"`
	Types      []string              `json:"types"`
	Attributes []SchemaAttribute     `json:"attributes"`
	Metadata   *SchemaMetadataUpdate `json:"metadata"`
}

// SchemaAttribute represents a schema attribute
type SchemaAttribute struct {
	Key         string                   `json:"key"`
	Name        map[string]string        `json:"name"`
	Description map[string]string        `json:"description,omitempty"`
	Type        string                   `json:"type"`
	Metadata    *SchemaAttributeMetadata `json:"metadata"`
	Values      []SchemaAttributeValue   `json:"values,omitempty"`
	Attributes  []SchemaAttribute        `json:"attributes,omitempty"`
	ArrayType   *SchemaArrayType         `json:"arrayType,omitempty"`
}

// SchemaAttributeMetadata represents metadata for a schema attribute
type SchemaAttributeMetadata struct {
	ReadOnly  bool `json:"readOnly"`
	Localized bool `json:"localized"`
	Required  bool `json:"required"`
	Nullable  bool `json:"nullable"`
}

// SchemaAttributeValue represents a value for ENUM/REFERENCE type attributes
type SchemaAttributeValue struct {
	Value string `json:"value"`
}

// SchemaArrayType represents the type configuration for ARRAY attributes
type SchemaArrayType struct {
	Type      string                 `json:"type"`
	Localized bool                   `json:"localized,omitempty"`
	Values    []SchemaAttributeValue `json:"values,omitempty"`
	// Attributes is required when Type == "OBJECT"
	Attributes []SchemaAttribute `json:"attributes,omitempty"`
}

// SchemaMetadata represents metadata for a schema
type SchemaMetadata struct {
	Version    int    `json:"version"`
	URL        string `json:"url,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	ModifiedAt string `json:"modifiedAt,omitempty"`
}

// SchemaMetadataUpdate represents metadata update for a schema
type SchemaMetadataUpdate struct {
	Version int `json:"version"`
}

// IdResponse represents a response containing just an ID
type IdResponse struct {
	ID string `json:"id"`
}

// CustomEntityOwner represents ownership information for a custom entity instance
type CustomEntityOwner struct {
	Type          string `json:"type"`
	UserID        string `json:"userId,omitempty"`
	LegalEntityID string `json:"legalEntityId,omitempty"`
}

// CustomEntityMetadata represents metadata for a custom entity instance
type CustomEntityMetadata struct {
	Version    int    `json:"version"`
	CreatedAt  string `json:"createdAt,omitempty"`
	ModifiedAt string `json:"modifiedAt,omitempty"`
}

// CustomEntityInstance represents a custom entity instance as returned by the API.
// "mixins" is a nested JSON object (confirmed against the live API - a Java backend
// deserializes it into a LinkedHashMap and rejects a JSON-encoded string value).
type CustomEntityInstance struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type,omitempty"`
	Name     map[string]string      `json:"name"`
	Owner    *CustomEntityOwner     `json:"owner,omitempty"`
	Mixins   map[string]interface{} `json:"mixins,omitempty"`
	Media    []string               `json:"media,omitempty"`
	Metadata *CustomEntityMetadata  `json:"metadata,omitempty"`
}

// CustomEntityInstanceCreate represents the creation payload for a custom entity instance
type CustomEntityInstanceCreate struct {
	ID     string                 `json:"id,omitempty"`
	Name   map[string]string      `json:"name"`
	Owner  *CustomEntityOwner     `json:"owner,omitempty"`
	Mixins map[string]interface{} `json:"mixins,omitempty"`
}

// CustomEntityInstanceUpdate represents the update payload for a custom entity instance.
// The API documents "id" as required on the PUT body (in addition to the URL path).
type CustomEntityInstanceUpdate struct {
	ID       string                 `json:"id"`
	Name     map[string]string      `json:"name"`
	Owner    *CustomEntityOwner     `json:"owner,omitempty"`
	Mixins   map[string]interface{} `json:"mixins,omitempty"`
	Metadata *SchemaMetadataUpdate  `json:"metadata,omitempty"`
}

// CustomEntityType represents a custom schema type definition
// (the "container" that custom entity instances belong to).
type CustomEntityType struct {
	ID       string                `json:"id"`
	Name     map[string]string     `json:"name"`
	Metadata *CustomEntityMetadata `json:"metadata,omitempty"`
}

// CustomEntityTypeCreate represents the creation payload for a custom schema type
type CustomEntityTypeCreate struct {
	ID   string            `json:"id"`
	Name map[string]string `json:"name"`
}

// CustomEntityTypeUpdate represents the update payload for a custom schema type
type CustomEntityTypeUpdate struct {
	Name     map[string]string     `json:"name"`
	Metadata *SchemaMetadataUpdate `json:"metadata,omitempty"`
}

// HeaderFieldValue represents a header value wrapper for the Emporix API.
// All header values must be wrapped in this struct with structure: {"value": "string"}
type HeaderFieldValue struct {
	Value string `json:"value"`
}

// EventConfig represents event-specific configuration for both read and write operations.
// Headers must use HeaderFieldValue wrapper. SecretKeyExists is used only in read responses.
type EventConfig struct {
	EventType       string                      `json:"eventType"`
	DestinationUrl  string                      `json:"destinationUrl,omitempty"`
	SecretKey       string                      `json:"secretKey,omitempty"`
	SecretKeyExists *bool                       `json:"secretKeyExists,omitempty"`
	Headers         map[string]HeaderFieldValue `json:"headers,omitempty"`
}

// WebhookConfig contains provider-specific configuration returned by the GET API.
type WebhookConfig struct {
	DestinationUrl      string                      `json:"destinationUrl,omitempty"`
	SecretKey           string                      `json:"secretKey,omitempty"`
	SecretKeyExists     *bool                       `json:"secretKeyExists,omitempty"`
	Headers             map[string]HeaderFieldValue `json:"headers,omitempty"`
	EventsConfiguration []EventConfig               `json:"eventsConfiguration,omitempty"`
}

// WebhookConfigGet represents the response for a webhook configuration.
type WebhookConfigGet struct {
	Code          string         `json:"code"`
	Active        bool           `json:"active"`
	Provider      string         `json:"provider"`
	Version       int            `json:"version"`
	Configuration *WebhookConfig `json:"configuration,omitempty"`
}

// NestedConfigCreate represents the nested configuration object for webhook creation.
// This contains provider-specific fields that the Emporix API expects.
// - For HTTP provider: uses `SecretKey` (JSON: `secretKey`) for HMAC signing.
// - For SVIX provider: uses `ApiKey` (JSON: `apiKey`) for Svix application secret.
type NestedConfigCreate struct {
	DestinationUrl      string                      `json:"destinationUrl,omitempty"`
	SecretKey           string                      `json:"secretKey,omitempty"`
	ApiKey              string                      `json:"apiKey,omitempty"`
	Headers             map[string]HeaderFieldValue `json:"headers,omitempty"`
	EventsConfiguration []EventConfig               `json:"eventsConfiguration,omitempty"`
}

// WebhookConfigPartialUpdates represents a JSON Patch operation for partial updates
type WebhookConfigPartialUpdates struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// WebhookListResponse represents the response for listing webhooks
type WebhookListResponse struct {
	Configs []WebhookConfigGet `json:"configs"`
}

// webhookCreateRequest is the creation payload for webhook configurations.
type webhookCreateRequest struct {
	Code          string              `json:"code"`
	Active        bool                `json:"active"`
	Provider      string              `json:"provider"`
	Configuration *NestedConfigCreate `json:"configuration,omitempty"`
}

type WebhookEventSubscriptionEntry struct {
	Event struct {
		Type string `json:"type"`
	} `json:"event"`
	Subscription string `json:"subscription"` // SUBSCRIBED | UNSUBSCRIBED | NONE
}

type WebhookEventSubscriptionUpdate struct {
	EventType string `json:"eventType"`
	Action    string `json:"action"` // SUBSCRIBE | UNSUBSCRIBE
}

type WebhookEventSubscriptionUpdateResult struct {
	EventType string `json:"eventType"`
	Code      int    `json:"code"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}
