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
