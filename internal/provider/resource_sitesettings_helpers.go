package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *SiteSettingsResource) terraformToApi(ctx context.Context, model *SiteSettingsResourceModel) (*SiteSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	site := &SiteSettings{
		Code:            model.Code.ValueString(),
		Name:            model.Name.ValueString(),
		Active:          model.Active.ValueBool(),
		Default:         model.Default.ValueBool(),
		DefaultLanguage: model.DefaultLanguage.ValueString(),
		Currency:        model.Currency.ValueString(),
	}

	// Optional bool with explicit handling
	if !model.IncludesTax.IsNull() {
		val := model.IncludesTax.ValueBool()
		site.IncludesTax = &val
	}

	// Languages
	var languages []string
	diags.Append(model.Languages.ElementsAs(ctx, &languages, false)...)
	site.Languages = languages

	// Available Currencies
	if !model.AvailableCurrencies.IsNull() {
		var availableCurrencies []string
		diags.Append(model.AvailableCurrencies.ElementsAs(ctx, &availableCurrencies, false)...)
		site.AvailableCurrencies = availableCurrencies
	}

	// Ship To Countries
	if !model.ShipToCountries.IsNull() {
		var shipToCountries []string
		diags.Append(model.ShipToCountries.ElementsAs(ctx, &shipToCountries, false)...)
		site.ShipToCountries = shipToCountries
	}

	// Tax Calculation Address Type
	if !model.TaxCalculationAddressType.IsNull() {
		site.TaxCalculationAddressType = model.TaxCalculationAddressType.ValueString()
	}

	// Decimal Points
	if !model.DecimalPoints.IsNull() {
		val := model.DecimalPoints.ValueInt64()
		site.DecimalPoints = &val
	}

	// Home Base
	if !model.HomeBase.IsNull() {
		homeBaseAttrs := model.HomeBase.Attributes()
		
		if len(homeBaseAttrs) > 0 {
			site.HomeBase = &HomeBase{}
			
			// Address
			if addressObj, ok := homeBaseAttrs["address"].(types.Object); ok && !addressObj.IsNull() {
				addressAttrs := addressObj.Attributes()
				
				site.HomeBase.Address = &Address{}
				if v, ok := addressAttrs["street"].(types.String); ok && !v.IsNull() {
					site.HomeBase.Address.Street = v.ValueString()
				}
				if v, ok := addressAttrs["street_number"].(types.String); ok && !v.IsNull() {
					site.HomeBase.Address.StreetNumber = v.ValueString()
				}
				if v, ok := addressAttrs["zip_code"].(types.String); ok && !v.IsNull() {
					site.HomeBase.Address.ZipCode = v.ValueString()
				}
				if v, ok := addressAttrs["city"].(types.String); ok && !v.IsNull() {
					site.HomeBase.Address.City = v.ValueString()
				}
				if v, ok := addressAttrs["country"].(types.String); ok && !v.IsNull() {
					site.HomeBase.Address.Country = v.ValueString()
				}
				if v, ok := addressAttrs["state"].(types.String); ok && !v.IsNull() {
					site.HomeBase.Address.State = v.ValueString()
				}
			}
			
			// Location
			if locationObj, ok := homeBaseAttrs["location"].(types.Object); ok && !locationObj.IsNull() {
				locationAttrs := locationObj.Attributes()
				
				site.HomeBase.Location = &Location{}
				if v, ok := locationAttrs["latitude"].(types.Float64); ok && !v.IsNull() {
					site.HomeBase.Location.Latitude = v.ValueFloat64()
				}
				if v, ok := locationAttrs["longitude"].(types.Float64); ok && !v.IsNull() {
					site.HomeBase.Location.Longitude = v.ValueFloat64()
				}
			}
		}
	}

	// Assisted Buying
	if !model.AssistedBuying.IsNull() {
		assistedBuyingAttrs := model.AssistedBuying.Attributes()
		
		if len(assistedBuyingAttrs) > 0 {
			site.AssistedBuying = &AssistedBuying{}
			if v, ok := assistedBuyingAttrs["storefront_url"].(types.String); ok && !v.IsNull() {
				site.AssistedBuying.StorefrontUrl = v.ValueString()
			}
		}
	}

	// Mixins
	if !model.Mixins.IsNull() {
		var mixins map[string]string
		diags.Append(model.Mixins.ElementsAs(ctx, &mixins, false)...)
		site.Mixins = mixins
	}

	return site, diags
}

func (r *SiteSettingsResource) apiToTerraform(ctx context.Context, site *SiteSettings, model *SiteSettingsResourceModel, previousModel *SiteSettingsResourceModel, diags *diag.Diagnostics) {
	model.Code = types.StringValue(site.Code)
	model.Name = types.StringValue(site.Name)
	model.Active = types.BoolValue(site.Active)
	model.Default = types.BoolValue(site.Default)
	model.DefaultLanguage = types.StringValue(site.DefaultLanguage)
	model.Currency = types.StringValue(site.Currency)

	// IncludesTax
	if site.IncludesTax != nil {
		model.IncludesTax = types.BoolValue(*site.IncludesTax)
	} else {
		model.IncludesTax = types.BoolNull()
	}

	// Languages
	if site.Languages != nil {
		languagesList, d := types.ListValueFrom(ctx, types.StringType, site.Languages)
		diags.Append(d...)
		model.Languages = languagesList
	}

	// Available Currencies
	if site.AvailableCurrencies != nil {
		currenciesList, d := types.ListValueFrom(ctx, types.StringType, site.AvailableCurrencies)
		diags.Append(d...)
		model.AvailableCurrencies = currenciesList
	} else {
		model.AvailableCurrencies = types.ListNull(types.StringType)
	}

	// Ship To Countries
	if site.ShipToCountries != nil {
		countriesList, d := types.ListValueFrom(ctx, types.StringType, site.ShipToCountries)
		diags.Append(d...)
		model.ShipToCountries = countriesList
	} else {
		model.ShipToCountries = types.ListNull(types.StringType)
	}

	// Tax Calculation Address Type - API doesn't reliably return this value, so preserve from previous state/plan
	if !previousModel.TaxCalculationAddressType.IsNull() && previousModel.TaxCalculationAddressType.ValueString() != "" {
		model.TaxCalculationAddressType = previousModel.TaxCalculationAddressType
	} else if site.TaxCalculationAddressType != "" {
		model.TaxCalculationAddressType = types.StringValue(site.TaxCalculationAddressType)
	} else {
		model.TaxCalculationAddressType = types.StringValue("BILLING_ADDRESS")
	}

	// Decimal Points
	if site.DecimalPoints != nil {
		model.DecimalPoints = types.Int64Value(*site.DecimalPoints)
	} else {
		model.DecimalPoints = types.Int64Value(2)
	}

	// Home Base
	if site.HomeBase != nil {
		homeBaseAttrs := make(map[string]attr.Value)
		
		// Address
		if site.HomeBase.Address != nil {
			addressAttrs := map[string]attr.Value{
				"street":        stringOrNull(site.HomeBase.Address.Street),
				"street_number": stringOrNull(site.HomeBase.Address.StreetNumber),
				"zip_code":      stringOrNull(site.HomeBase.Address.ZipCode),
				"city":          stringOrNull(site.HomeBase.Address.City),
				"country":       types.StringValue(site.HomeBase.Address.Country),
				"state":         stringOrNull(site.HomeBase.Address.State),
			}
			addressObj, d := types.ObjectValue(map[string]attr.Type{
				"street":        types.StringType,
				"street_number": types.StringType,
				"zip_code":      types.StringType,
				"city":          types.StringType,
				"country":       types.StringType,
				"state":         types.StringType,
			}, addressAttrs)
			diags.Append(d...)
			homeBaseAttrs["address"] = addressObj
		} else {
			homeBaseAttrs["address"] = types.ObjectNull(map[string]attr.Type{
				"street":        types.StringType,
				"street_number": types.StringType,
				"zip_code":      types.StringType,
				"city":          types.StringType,
				"country":       types.StringType,
				"state":         types.StringType,
			})
		}
		
		// Location
		if site.HomeBase.Location != nil {
			locationAttrs := map[string]attr.Value{
				"latitude":  types.Float64Value(site.HomeBase.Location.Latitude),
				"longitude": types.Float64Value(site.HomeBase.Location.Longitude),
			}
			locationObj, d := types.ObjectValue(map[string]attr.Type{
				"latitude":  types.Float64Type,
				"longitude": types.Float64Type,
			}, locationAttrs)
			diags.Append(d...)
			homeBaseAttrs["location"] = locationObj
		} else {
			homeBaseAttrs["location"] = types.ObjectNull(map[string]attr.Type{
				"latitude":  types.Float64Type,
				"longitude": types.Float64Type,
			})
		}
		
		homeBaseObj, d := types.ObjectValue(map[string]attr.Type{
			"address": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"street":        types.StringType,
					"street_number": types.StringType,
					"zip_code":      types.StringType,
					"city":          types.StringType,
					"country":       types.StringType,
					"state":         types.StringType,
				},
			},
			"location": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"latitude":  types.Float64Type,
					"longitude": types.Float64Type,
				},
			},
		}, homeBaseAttrs)
		diags.Append(d...)
		model.HomeBase = homeBaseObj
	} else {
		model.HomeBase = types.ObjectNull(map[string]attr.Type{
			"address": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"street":        types.StringType,
					"street_number": types.StringType,
					"zip_code":      types.StringType,
					"city":          types.StringType,
					"country":       types.StringType,
					"state":         types.StringType,
				},
			},
			"location": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"latitude":  types.Float64Type,
					"longitude": types.Float64Type,
				},
			},
		})
	}

	// Assisted Buying
	if site.AssistedBuying != nil {
		assistedBuyingAttrs := map[string]attr.Value{
			"storefront_url": stringOrNull(site.AssistedBuying.StorefrontUrl),
		}
		assistedBuyingObj, d := types.ObjectValue(map[string]attr.Type{
			"storefront_url": types.StringType,
		}, assistedBuyingAttrs)
		diags.Append(d...)
		model.AssistedBuying = assistedBuyingObj
	} else {
		model.AssistedBuying = types.ObjectNull(map[string]attr.Type{
			"storefront_url": types.StringType,
		})
	}

	// Mixins
	if site.Mixins != nil && len(site.Mixins) > 0 {
		mixinsMap, d := types.MapValueFrom(ctx, types.StringType, site.Mixins)
		diags.Append(d...)
		model.Mixins = mixinsMap
	} else {
		model.Mixins = types.MapNull(types.StringType)
	}
}

// Helper function to convert empty strings to null
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
