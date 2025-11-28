package provider

import (
	"context"
	"encoding/json"

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

	// Cart Calculation Scale
	if !model.CartCalculationScale.IsNull() {
		val := model.CartCalculationScale.ValueInt64()
		site.CartCalculationScale = &val
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

	// Metadata
	if !model.Metadata.IsNull() {
		metadataAttrs := model.Metadata.Attributes()
		
		if len(metadataAttrs) > 0 {
			site.Metadata = &Metadata{}
			
			// Mixins URLs map (version is managed by API)
			if mixinsMapObj, ok := metadataAttrs["mixins"].(types.Map); ok && !mixinsMapObj.IsNull() {
				var mixinsMap map[string]string
				diagsTemp := mixinsMapObj.ElementsAs(ctx, &mixinsMap, false)
				diags.Append(diagsTemp...)
				if len(mixinsMap) > 0 {
					site.Metadata.Mixins = mixinsMap
				}
			}
		}
	}

	// Mixins (JSON string)
	if !model.Mixins.IsNull() && model.Mixins.ValueString() != "" {
		var mixinsData map[string]interface{}
		if err := json.Unmarshal([]byte(model.Mixins.ValueString()), &mixinsData); err != nil {
			diags.AddError(
				"Invalid Mixins JSON",
				"Failed to parse mixins JSON: "+err.Error(),
			)
		} else {
			site.Mixins = mixinsData
		}
	}

	return site, diags
}

// buildPatchData creates a map with only the fields that changed between state and plan
func (r *SiteSettingsResource) buildPatchData(ctx context.Context, plan *SiteSettingsResourceModel, state *SiteSettingsResourceModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	patchData := make(map[string]interface{})

	// Name
	if !plan.Name.Equal(state.Name) {
		patchData["name"] = plan.Name.ValueString()
	}

	// Active
	if !plan.Active.Equal(state.Active) {
		patchData["active"] = plan.Active.ValueBool()
	}

	// Default
	if !plan.Default.Equal(state.Default) {
		patchData["default"] = plan.Default.ValueBool()
	}

	// IncludesTax
	if !plan.IncludesTax.Equal(state.IncludesTax) {
		if !plan.IncludesTax.IsNull() {
			patchData["includesTax"] = plan.IncludesTax.ValueBool()
		} else {
			patchData["includesTax"] = nil
		}
	}

	// DefaultLanguage
	if !plan.DefaultLanguage.Equal(state.DefaultLanguage) {
		patchData["defaultLanguage"] = plan.DefaultLanguage.ValueString()
	}

	// Languages
	if !plan.Languages.Equal(state.Languages) {
		var languages []string
		diags.Append(plan.Languages.ElementsAs(ctx, &languages, false)...)
		patchData["languages"] = languages
	}

	// Currency
	if !plan.Currency.Equal(state.Currency) {
		patchData["currency"] = plan.Currency.ValueString()
	}

	// AvailableCurrencies
	if !plan.AvailableCurrencies.Equal(state.AvailableCurrencies) {
		if !plan.AvailableCurrencies.IsNull() {
			var availableCurrencies []string
			diags.Append(plan.AvailableCurrencies.ElementsAs(ctx, &availableCurrencies, false)...)
			patchData["availableCurrencies"] = availableCurrencies
		} else {
			patchData["availableCurrencies"] = nil
		}
	}

	// ShipToCountries
	if !plan.ShipToCountries.Equal(state.ShipToCountries) {
		if !plan.ShipToCountries.IsNull() {
			var shipToCountries []string
			diags.Append(plan.ShipToCountries.ElementsAs(ctx, &shipToCountries, false)...)
			patchData["shipToCountries"] = shipToCountries
		} else {
			patchData["shipToCountries"] = nil
		}
	}

	// TaxCalculationAddressType
	if !plan.TaxCalculationAddressType.Equal(state.TaxCalculationAddressType) {
		if !plan.TaxCalculationAddressType.IsNull() {
			patchData["taxCalculationAddressType"] = plan.TaxCalculationAddressType.ValueString()
		} else {
			patchData["taxCalculationAddressType"] = nil
		}
	}

	// DecimalPoints
	if !plan.DecimalPoints.Equal(state.DecimalPoints) {
		if !plan.DecimalPoints.IsNull() {
			patchData["decimalPoints"] = plan.DecimalPoints.ValueInt64()
		} else {
			patchData["decimalPoints"] = nil
		}
	}

	// CartCalculationScale
	if !plan.CartCalculationScale.Equal(state.CartCalculationScale) {
		if !plan.CartCalculationScale.IsNull() {
			patchData["cartCalculationScale"] = plan.CartCalculationScale.ValueInt64()
		} else {
			patchData["cartCalculationScale"] = nil
		}
	}

	// HomeBase
	if !plan.HomeBase.Equal(state.HomeBase) {
		if !plan.HomeBase.IsNull() {
			homeBaseAttrs := plan.HomeBase.Attributes()
			homeBase := make(map[string]interface{})
			
			// Address
			if addressObj, ok := homeBaseAttrs["address"].(types.Object); ok && !addressObj.IsNull() {
				addressAttrs := addressObj.Attributes()
				address := make(map[string]interface{})
				
				if v, ok := addressAttrs["street"].(types.String); ok && !v.IsNull() {
					address["street"] = v.ValueString()
				}
				if v, ok := addressAttrs["street_number"].(types.String); ok && !v.IsNull() {
					address["streetNumber"] = v.ValueString()
				}
				if v, ok := addressAttrs["zip_code"].(types.String); ok && !v.IsNull() {
					address["zipCode"] = v.ValueString()
				}
				if v, ok := addressAttrs["city"].(types.String); ok && !v.IsNull() {
					address["city"] = v.ValueString()
				}
				if v, ok := addressAttrs["country"].(types.String); ok && !v.IsNull() {
					address["country"] = v.ValueString()
				}
				if v, ok := addressAttrs["state"].(types.String); ok && !v.IsNull() {
					address["state"] = v.ValueString()
				}
				
				if len(address) > 0 {
					homeBase["address"] = address
				}
			}
			
			// Location
			if locationObj, ok := homeBaseAttrs["location"].(types.Object); ok && !locationObj.IsNull() {
				locationAttrs := locationObj.Attributes()
				location := make(map[string]interface{})
				
				if v, ok := locationAttrs["latitude"].(types.Float64); ok && !v.IsNull() {
					location["latitude"] = v.ValueFloat64()
				}
				if v, ok := locationAttrs["longitude"].(types.Float64); ok && !v.IsNull() {
					location["longitude"] = v.ValueFloat64()
				}
				
				if len(location) > 0 {
					homeBase["location"] = location
				}
			}
			
			if len(homeBase) > 0 {
				patchData["homeBase"] = homeBase
			}
		} else {
			patchData["homeBase"] = nil
		}
	}

	// AssistedBuying
	if !plan.AssistedBuying.Equal(state.AssistedBuying) {
		if !plan.AssistedBuying.IsNull() {
			assistedBuyingAttrs := plan.AssistedBuying.Attributes()
			assistedBuying := make(map[string]interface{})
			
			if v, ok := assistedBuyingAttrs["storefront_url"].(types.String); ok && !v.IsNull() {
				assistedBuying["storefrontUrl"] = v.ValueString()
			}
			
			if len(assistedBuying) > 0 {
				patchData["assistedBuying"] = assistedBuying
			}
		} else {
			patchData["assistedBuying"] = nil
		}
	}

	return patchData, diags
}

func (r *SiteSettingsResource) apiToTerraform(ctx context.Context, site *SiteSettings, model *SiteSettingsResourceModel, previousModel *SiteSettingsResourceModel, diags *diag.Diagnostics, preservePlanValues bool) {
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

	// Available Currencies - API may return this even if not set
	// During Read: show actual values (detect drift)
	// During Create/Update: preserve plan values (prevent consistency errors)
	if !preservePlanValues || !previousModel.AvailableCurrencies.IsNull() {
		// Either in Read mode, or user specified available_currencies
		if site.AvailableCurrencies != nil {
			currenciesList, d := types.ListValueFrom(ctx, types.StringType, site.AvailableCurrencies)
			diags.Append(d...)
			model.AvailableCurrencies = currenciesList
		} else {
			model.AvailableCurrencies = types.ListNull(types.StringType)
		}
	} else {
		// In Create/Update mode and user didn't specify, keep null
		model.AvailableCurrencies = types.ListNull(types.StringType)
	}

	// Ship To Countries - API may return this even if not set
	// During Read: show actual values (detect drift)
	// During Create/Update: preserve plan values (prevent consistency errors)
	if !preservePlanValues || !previousModel.ShipToCountries.IsNull() {
		// Either in Read mode, or user specified ship_to_countries
		if site.ShipToCountries != nil {
			countriesList, d := types.ListValueFrom(ctx, types.StringType, site.ShipToCountries)
			diags.Append(d...)
			model.ShipToCountries = countriesList
		} else {
			model.ShipToCountries = types.ListNull(types.StringType)
		}
	} else {
		// In Create/Update mode and user didn't specify, keep null
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

	// Cart Calculation Scale
	if site.CartCalculationScale != nil {
		model.CartCalculationScale = types.Int64Value(*site.CartCalculationScale)
	} else {
		model.CartCalculationScale = types.Int64Value(2)
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

	// Metadata - API may return this even if not set (version is managed by API, not exposed to users)
	// During Read: show actual values (detect drift)
	// During Create/Update: preserve plan values (prevent consistency errors)
	if !preservePlanValues || !previousModel.Metadata.IsNull() {
		// Either in Read mode, or user specified metadata
		if site.Metadata != nil && len(site.Metadata.Mixins) > 0 {
			metadataAttrs := make(map[string]attr.Value)
			
			// Mixins URLs map (only field we expose to users)
			if site.Metadata.Mixins != nil && len(site.Metadata.Mixins) > 0 {
				mixinsMap, d := types.MapValueFrom(ctx, types.StringType, site.Metadata.Mixins)
				diags.Append(d...)
				metadataAttrs["mixins"] = mixinsMap
			} else {
				metadataAttrs["mixins"] = types.MapNull(types.StringType)
			}
			
			metadataObj, d := types.ObjectValue(map[string]attr.Type{
				"mixins": types.MapType{ElemType: types.StringType},
			}, metadataAttrs)
			diags.Append(d...)
			model.Metadata = metadataObj
		} else {
			model.Metadata = types.ObjectNull(map[string]attr.Type{
				"mixins": types.MapType{ElemType: types.StringType},
			})
		}
	} else {
		// In Create/Update mode and user didn't specify metadata, keep null
		model.Metadata = types.ObjectNull(map[string]attr.Type{
			"mixins": types.MapType{ElemType: types.StringType},
		})
	}

	// Mixins (convert to JSON string) - API may return this even if not set
	// During Read: show actual values (detect drift)
	// During Create/Update: preserve plan values (prevent consistency errors)
	if !preservePlanValues || (!previousModel.Mixins.IsNull() && previousModel.Mixins.ValueString() != "") {
		// Either in Read mode, or user specified mixins
		if site.Mixins != nil && len(site.Mixins) > 0 {
			mixinsJSON, err := json.Marshal(site.Mixins)
			if err != nil {
				diags.AddError(
					"Failed to Marshal Mixins",
					"Could not convert mixins to JSON: "+err.Error(),
				)
				model.Mixins = types.StringNull()
			} else {
				model.Mixins = types.StringValue(string(mixinsJSON))
			}
		} else {
			model.Mixins = types.StringNull()
		}
	} else {
		// In Create/Update mode and user didn't specify mixins, keep null
		model.Mixins = types.StringNull()
	}
}

// Helper function to convert empty strings to null
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
