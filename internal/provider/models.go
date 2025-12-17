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
