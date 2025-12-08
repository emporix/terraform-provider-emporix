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
	Version int                `json:"version,omitempty"`
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
