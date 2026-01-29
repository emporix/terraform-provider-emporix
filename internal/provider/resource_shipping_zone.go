package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ShippingZoneResource{}
	_ resource.ResourceWithImportState = &ShippingZoneResource{}
)

type ShippingZoneResource struct {
	client *EmporixClient
}

type ShippingZoneResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Site    types.String `tfsdk:"site"`
	Name    types.Map    `tfsdk:"name"`
	Default types.Bool   `tfsdk:"default"`
	ShipTo  types.Set    `tfsdk:"ship_to"`
}

type ShippingDestinationModel struct {
	Country    types.String `tfsdk:"country"`
	PostalCode types.String `tfsdk:"postal_code"`
}

// API structs for ShippingZone

// ShippingZone represents a shipping zone
type ShippingZone struct {
	ID      string                `json:"id"`
	Name    interface{}           `json:"name"`
	Default bool                  `json:"default,omitempty"`
	ShipTo  []ShippingDestination `json:"shipTo"`
}

// ShippingDestination represents a shipping destination
type ShippingDestination struct {
	Country    string `json:"country"`
	PostalCode string `json:"postalCode,omitempty"`
}

func NewShippingZoneResource() resource.Resource {
	return &ShippingZoneResource{}
}

func (r *ShippingZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shipping_zone"
}

func (r *ShippingZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages shipping zones for a site.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Shipping zone's unique identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site identifier. Typically 'main' for single-shop tenants.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.MapAttribute{
				MarkdownDescription: "Zone name as a map of language codes to translated names. Example: `{en = \"Zone Name\"}` for single language or `{en = \"English\", de = \"Deutsch\"}` for multiple languages. The map is always sent to the API as-is.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"default": schema.BoolAttribute{
				MarkdownDescription: "Flag indicating whether the zone is the default delivery zone for the site. If not specified, the API may automatically set this to true for the first zone.",
				Optional:            true,
				Computed:            true,
			},
			"ship_to": schema.SetNestedAttribute{
				MarkdownDescription: "Collection of shipping destinations. Each country can only appear once in the list. The order of destinations does not matter as the API stores them in sorted order.",
				Required:            true,
				Validators: []validator.Set{
					uniqueCountryValidatorSet{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"country": schema.StringAttribute{
							MarkdownDescription: "Country code (e.g., 'DE', 'US').",
							Required:            true,
						},
						"postal_code": schema.StringAttribute{
							MarkdownDescription: "Postal code or postal code pattern.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *ShippingZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// ProviderData will be nil when the resource is initially created
	// This is expected behavior during the configure phase
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*EmporixClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *EmporixClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ShippingZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ShippingZoneResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating shipping zone", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"site": data.Site.ValueString(),
	})

	// Parse name map
	var nameMap map[string]string
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ship_to destinations
	var shipToModels []ShippingDestinationModel
	resp.Diagnostics.Append(data.ShipTo.ElementsAs(ctx, &shipToModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	shipTo := make([]ShippingDestination, len(shipToModels))
	for i, dest := range shipToModels {
		shipTo[i] = ShippingDestination{
			Country:    dest.Country.ValueString(),
			PostalCode: dest.PostalCode.ValueString(),
		}
	}

	zone := &ShippingZone{
		ID:      data.ID.ValueString(),
		Name:    nameMap,
		Default: data.Default.ValueBool(),
		ShipTo:  shipTo,
	}

	// Create zone via API
	_, err := r.client.CreateShippingZone(ctx, data.Site.ValueString(), zone)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create shipping zone, got error: %s", err))
		return
	}

	// Read back the created resource to get the actual state from the API
	// This ensures all fields match what's actually stored, including computed values
	tflog.Debug(ctx, "Reading back created shipping zone to ensure state consistency")
	actualZone, err := r.client.GetShippingZone(ctx, data.Site.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created shipping zone, got error: %s", err))
		return
	}

	// Sync ALL fields to state
	// Note: site and id are preserved from plan (site is path parameter, not in API response)
	// data.Site already set from plan
	data.ID = types.StringValue(actualZone.ID)
	data.Default = types.BoolValue(actualZone.Default)

	// Convert API response name - use actual response, not original nameMap
	if actualZone.Name != nil {
		nameMapResult, diagsName := convertNameToMap(ctx, actualZone.Name, nameMap)
		resp.Diagnostics.Append(diagsName...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Name = nameMapResult
	}

	// Convert ship_to from actual response
	if len(actualZone.ShipTo) > 0 {
		var diagsShipTo diag.Diagnostics
		data.ShipTo, diagsShipTo = convertShipToToSet(ctx, actualZone.ShipTo)
		resp.Diagnostics.Append(diagsShipTo...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShippingZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ShippingZoneResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading shipping zone", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"site": data.Site.ValueString(),
	})

	// Get current state from API
	actualZone, err := r.client.GetShippingZone(ctx, data.Site.ValueString(), data.ID.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read shipping zone, got error: %s", err))
		return
	}

	// Get the original name map from state to preserve language keys
	var originalNameMap map[string]string
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &originalNameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sync ALL fields to state from the API response
	// Note: site is preserved from existing state (it's a path parameter, not in API response)
	// data.Site already set from state
	data.ID = types.StringValue(actualZone.ID)
	data.Default = types.BoolValue(actualZone.Default)

	// Convert API response name
	if actualZone.Name != nil {
		nameMapResult, diagsName := convertNameToMap(ctx, actualZone.Name, originalNameMap)
		resp.Diagnostics.Append(diagsName...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Name = nameMapResult
	}

	// Convert ship_to from API response
	if len(actualZone.ShipTo) > 0 {
		var diagsShipTo diag.Diagnostics
		data.ShipTo, diagsShipTo = convertShipToToSet(ctx, actualZone.ShipTo)
		resp.Diagnostics.Append(diagsShipTo...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShippingZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ShippingZoneResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating shipping zone", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"site": data.Site.ValueString(),
	})

	// Parse name map
	var nameMap map[string]string
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ship_to destinations
	var shipToModels []ShippingDestinationModel
	resp.Diagnostics.Append(data.ShipTo.ElementsAs(ctx, &shipToModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	shipTo := make([]ShippingDestination, len(shipToModels))
	for i, dest := range shipToModels {
		shipTo[i] = ShippingDestination{
			Country:    dest.Country.ValueString(),
			PostalCode: dest.PostalCode.ValueString(),
		}
	}

	zone := &ShippingZone{
		ID:      data.ID.ValueString(),
		Name:    nameMap,
		Default: data.Default.ValueBool(),
		ShipTo:  shipTo,
	}

	// Update zone via API
	_, err := r.client.UpdateShippingZone(ctx, data.Site.ValueString(), data.ID.ValueString(), zone)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update shipping zone, got error: %s", err))
		return
	}

	// Read back the updated resource to get the actual state from the API
	// This ensures all fields match what's actually stored, including computed values
	tflog.Debug(ctx, "Reading back updated shipping zone to ensure state consistency")
	actualZone, err := r.client.GetShippingZone(ctx, data.Site.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated shipping zone, got error: %s", err))
		return
	}

	// Sync ALL fields to state
	// Note: site and id are preserved from plan (site is path parameter, not in API response)
	// data.Site already set from plan
	data.ID = types.StringValue(actualZone.ID)
	data.Default = types.BoolValue(actualZone.Default)

	// Convert API response name - use actual response, not original nameMap
	if actualZone.Name != nil {
		nameMapResult, diagsName := convertNameToMap(ctx, actualZone.Name, nameMap)
		resp.Diagnostics.Append(diagsName...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Name = nameMapResult
	}

	// Convert ship_to from actual response
	if len(actualZone.ShipTo) > 0 {
		var diagsShipTo diag.Diagnostics
		data.ShipTo, diagsShipTo = convertShipToToSet(ctx, actualZone.ShipTo)
		resp.Diagnostics.Append(diagsShipTo...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShippingZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ShippingZoneResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting shipping zone", map[string]interface{}{
		"id":      data.ID.ValueString(),
		"site":    data.Site.ValueString(),
		"default": data.Default.ValueBool(),
	})

	// If this is the default zone, we need to handle it specially
	// The API won't let us delete the default zone unless it's the only one
	if data.Default.ValueBool() {
		tflog.Debug(ctx, "Zone is default, checking if we need to reassign default to another zone")

		// List all zones to see if there are others
		allZones, err := r.client.ListShippingZones(ctx, data.Site.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list shipping zones, got error: %s", err))
			return
		}

		// Find another zone to make default (any zone that's not the one we're deleting)
		var otherZone *ShippingZone
		for i := range allZones {
			if allZones[i].ID != data.ID.ValueString() {
				otherZone = &allZones[i]
				break
			}
		}

		// If there's another zone, make it the default first
		if otherZone != nil {
			tflog.Debug(ctx, "Making another zone default before deletion", map[string]interface{}{
				"new_default_zone_id": otherZone.ID,
			})

			// Update the other zone to be default
			otherZone.Default = true
			_, err := r.client.UpdateShippingZone(ctx, data.Site.ValueString(), otherZone.ID, otherZone)
			if err != nil {
				resp.Diagnostics.AddError("Client Error",
					fmt.Sprintf("Unable to reassign default zone before deletion, got error: %s", err))
				return
			}

			tflog.Debug(ctx, "Successfully reassigned default to another zone")
		} else {
			tflog.Debug(ctx, "This is the only zone, deletion should succeed")
		}
	}

	// Now delete the zone
	err := r.client.DeleteShippingZone(ctx, data.Site.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete shipping zone, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "Successfully deleted shipping zone")
}

func (r *ShippingZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: "site:zone-id"
	// Example: "main:zone-express"
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'site:zone-id', got: %s", req.ID),
		)
		return
	}

	site := parts[0]
	zoneID := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), zoneID)...)
}

// uniqueCountryValidatorSet validates that each country appears only once in ship_to set
type uniqueCountryValidatorSet struct{}

func (v uniqueCountryValidatorSet) Description(ctx context.Context) string {
	return "Ensures each country code appears only once in the ship_to set"
}

func (v uniqueCountryValidatorSet) MarkdownDescription(ctx context.Context) string {
	return "Ensures each country code appears only once in the ship_to set. The Emporix API does not allow multiple entries for the same country with different postal codes."
}

func (v uniqueCountryValidatorSet) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var destinations []ShippingDestinationModel
	diags := req.ConfigValue.ElementsAs(ctx, &destinations, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Track countries we've seen
	seenCountries := make(map[string]bool)
	duplicates := make(map[string]bool)

	for _, dest := range destinations {
		if dest.Country.IsNull() || dest.Country.IsUnknown() {
			continue
		}

		country := dest.Country.ValueString()
		if seenCountries[country] {
			duplicates[country] = true
		}
		seenCountries[country] = true
	}

	// Report all duplicate countries
	if len(duplicates) > 0 {
		for country := range duplicates {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate Country in ship_to",
				fmt.Sprintf("The country code '%s' appears multiple times in ship_to. "+
					"The Emporix API does not allow multiple entries for the same country. "+
					"Each country can only appear once in the ship_to set.", country),
			)
		}
	}
}

// uniqueCountryValidator validates that each country appears only once in ship_to list (deprecated, kept for compatibility)
// Helper function to convert ShippingDestination slice to types.Set
func convertShipToToSet(ctx context.Context, shipTo []ShippingDestination) (types.Set, diag.Diagnostics) {
	// No need to sort - sets are unordered!
	var destinations []ShippingDestinationModel
	for _, dest := range shipTo {
		country := types.StringValue(dest.Country)
		postalCode := types.StringValue(dest.PostalCode)

		// Use null for empty values instead of empty strings
		if dest.Country == "" {
			country = types.StringNull()
		}
		if dest.PostalCode == "" {
			postalCode = types.StringNull()
		}

		destinations = append(destinations, ShippingDestinationModel{
			Country:    country,
			PostalCode: postalCode,
		})
	}

	setValue, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"country":     types.StringType,
			"postal_code": types.StringType,
		},
	}, destinations)

	return setValue, diags
}

// Helper function to convert API name response to types.Map
func convertNameToMap(ctx context.Context, apiName interface{}, originalMap map[string]string) (types.Map, diag.Diagnostics) {
	if apiName == nil {
		// API didn't return name, use original
		return types.MapValueFrom(ctx, types.StringType, originalMap)
	}

	// Try to convert API response to map
	resultMap := make(map[string]string)

	switch v := apiName.(type) {
	case string:
		// API returned a simple string - use the original map keys to preserve the language
		if len(originalMap) == 1 {
			for lang := range originalMap {
				resultMap[lang] = v
			}
		} else {
			// This shouldn't happen, but fallback to using the original map
			resultMap = originalMap
		}
	case map[string]interface{}:
		// API returned a map - convert it
		for lang, val := range v {
			if strVal, ok := val.(string); ok {
				resultMap[lang] = strVal
			}
		}
	default:
		// Unknown type, use original
		resultMap = originalMap
	}

	return types.MapValueFrom(ctx, types.StringType, resultMap)
}
