package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &DeliveryTimeResource{}
	_ resource.ResourceWithConfigure   = &DeliveryTimeResource{}
	_ resource.ResourceWithImportState = &DeliveryTimeResource{}
)

func NewDeliveryTimeResource() resource.Resource {
	return &DeliveryTimeResource{}
}

type DeliveryTimeResource struct {
	client *EmporixClient
}

type DeliveryTimeResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	SiteCode         types.String `tfsdk:"site_code"`
	IsDeliveryDay    types.Bool   `tfsdk:"is_delivery_day"`
	ZoneID           types.String `tfsdk:"zone_id"`
	Day              types.Object `tfsdk:"day"`
	IsForAllZones    types.Bool   `tfsdk:"is_for_all_zones"`
	TimeZoneID       types.String `tfsdk:"time_zone_id"`
	DeliveryDayShift types.Int64  `tfsdk:"delivery_day_shift"`
	Slots            types.List   `tfsdk:"slots"`
}

type DeliveryDayModel struct {
	Weekday  types.String `tfsdk:"weekday"`
	Date     types.String `tfsdk:"date"`
	DateFrom types.String `tfsdk:"date_from"`
	DateTo   types.String `tfsdk:"date_to"`
}

type DeliveryTimeSlotModel struct {
	ShippingMethod    types.String `tfsdk:"shipping_method"`
	DeliveryTimeRange types.Object `tfsdk:"delivery_time_range"`
	CutOffTime        types.Object `tfsdk:"cut_off_time"`
	Capacity          types.Int64  `tfsdk:"capacity"`
}

type TimeRangeModel struct {
	TimeFrom types.String `tfsdk:"time_from"`
	TimeTo   types.String `tfsdk:"time_to"`
}

type CutOffTimeModel struct {
	Time              types.String `tfsdk:"time"`
	DeliveryCycleName types.String `tfsdk:"delivery_cycle_name"`
}

// API structs for DeliveryTime

// DeliveryTime represents a delivery time configuration
type DeliveryTime struct {
	ID               string             `json:"id,omitempty"`
	SiteCode         string             `json:"siteCode"`
	Name             string             `json:"name"`
	IsDeliveryDay    bool               `json:"isDeliveryDay"`
	ZoneID           string             `json:"zoneId,omitempty"`
	Day              *DeliveryDay       `json:"day,omitempty"`
	IsForAllZones    bool               `json:"isForAllZones"`
	TimeZoneID       string             `json:"timeZoneId"`
	DeliveryDayShift int                `json:"deliveryDayShift"`
	Slots            []DeliveryTimeSlot `json:"slots,omitempty"`
}

// DeliveryDay represents the day configuration
// DatePeriod represents a date range for delivery times
type DatePeriod struct {
	DateFrom string `json:"dateFrom,omitempty"` // Start date
	DateTo   string `json:"dateTo,omitempty"`   // End date
}

// DeliveryDay represents the day configuration for delivery times
type DeliveryDay struct {
	Weekday    string      `json:"weekday,omitempty"`    // MONDAY, TUESDAY, etc.
	SingleDate string      `json:"singleDate,omitempty"` // Specific date in ISO 8601 with time
	DatePeriod *DatePeriod `json:"datePeriod,omitempty"` // Date range object
}

// DeliveryTimeSlot represents a delivery time slot
type DeliveryTimeSlot struct {
	ShippingMethod    string      `json:"shippingMethod"`
	DeliveryTimeRange *TimeRange  `json:"deliveryTimeRange"`
	CutOffTime        *CutOffTime `json:"cutOffTime,omitempty"`
	Capacity          int         `json:"capacity"`
}

// TimeRange represents a time range
type TimeRange struct {
	TimeFrom string `json:"timeFrom"` // HH:MM format
	TimeTo   string `json:"timeTo"`   // HH:MM format
}

// CutOffTime represents the cutoff time configuration
type CutOffTime struct {
	Time              string `json:"time"` // ISO 8601 format
	DeliveryCycleName string `json:"deliveryCycleName"`
}

func (r *DeliveryTimeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_time"
}

func (r *DeliveryTimeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages delivery time slot configurations for scheduled deliveries.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier generated by the API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique name for the delivery time configuration.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"site_code": schema.StringAttribute{
				MarkdownDescription: "Site code. Typically 'main' for single-shop tenants.",
				Required:            true,
			},
			"is_delivery_day": schema.BoolAttribute{
				MarkdownDescription: "Whether this is an active delivery day.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"zone_id": schema.StringAttribute{
				MarkdownDescription: "Shipping zone ID this delivery time applies to. Required if is_for_all_zones is false.",
				Optional:            true,
			},
			"is_for_all_zones": schema.BoolAttribute{
				MarkdownDescription: "Whether this delivery time applies to all zones.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"time_zone_id": schema.StringAttribute{
				MarkdownDescription: "Timezone identifier (e.g., 'Europe/Warsaw', 'America/New_York').",
				Required:            true,
			},
			"delivery_day_shift": schema.Int64Attribute{
				MarkdownDescription: "Number of days to shift delivery from order date.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"day": schema.SingleNestedAttribute{
				MarkdownDescription: "Day configuration for this delivery time. Supports weekday (recurring), specific date, or date range.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"weekday": schema.StringAttribute{
						MarkdownDescription: "Day of the week for recurring delivery (MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY). Use for weekly recurring deliveries.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"),
						},
					},
					"date": schema.StringAttribute{
						MarkdownDescription: "Specific date for one-time delivery in ISO 8601 format. Example: '2024-12-25T10:00:00.000Z'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`),
								"must be in ISO 8601 format (YYYY-MM-DDTHH:MM:SS.sssZ)",
							),
						},
					},
					"date_from": schema.StringAttribute{
						MarkdownDescription: "Start date for delivery date range in ISO 8601 format. Example: '2024-06-01T10:00:00.000Z'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`),
								"must be in ISO 8601 format (YYYY-MM-DDTHH:MM:SS.sssZ)",
							),
						},
					},
					"date_to": schema.StringAttribute{
						MarkdownDescription: "End date for delivery date range in ISO 8601 format. Example: '2024-08-31T10:00:00.000Z'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`),
								"must be in ISO 8601 format (YYYY-MM-DDTHH:MM:SS.sssZ)",
							),
						},
					},
				},
			},
			"slots": schema.ListNestedAttribute{
				MarkdownDescription: "Delivery time slots with shipping methods and capacity.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"shipping_method": schema.StringAttribute{
							MarkdownDescription: "Shipping method identifier.",
							Required:            true,
						},
						"capacity": schema.Int64Attribute{
							MarkdownDescription: "Maximum number of deliveries for this slot.",
							Required:            true,
						},
						"delivery_time_range": schema.SingleNestedAttribute{
							MarkdownDescription: "Time range for delivery.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"time_from": schema.StringAttribute{
									MarkdownDescription: "Start time in HH:MM format (e.g., '10:00').",
									Required:            true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(
											regexp.MustCompile(`^([0-1][0-9]|2[0-3]):([0-5][0-9])$`),
											"must be in HH:MM format (e.g., '10:00')",
										),
									},
								},
								"time_to": schema.StringAttribute{
									MarkdownDescription: "End time in HH:MM format (e.g., '12:00').",
									Required:            true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(
											regexp.MustCompile(`^([0-1][0-9]|2[0-3]):([0-5][0-9])$`),
											"must be in HH:MM format (e.g., '12:00')",
										),
									},
								},
							},
						},
						"cut_off_time": schema.SingleNestedAttribute{
							MarkdownDescription: "Order cutoff time for this slot.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"time": schema.StringAttribute{
									MarkdownDescription: "Cutoff timestamp in ISO 8601 format (e.g., '2023-06-12T18:00:00.000Z').",
									Required:            true,
								},
								"delivery_cycle_name": schema.StringAttribute{
									MarkdownDescription: "Delivery cycle identifier (e.g., 'morning', 'afternoon').",
									Required:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *DeliveryTimeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// buildDeliveryTimeFromModel builds a DeliveryTime API struct from the Terraform model
func buildDeliveryTimeFromModel(ctx context.Context, data *DeliveryTimeResourceModel, diags *diag.Diagnostics) *DeliveryTime {
	deliveryTime := &DeliveryTime{
		SiteCode:         data.SiteCode.ValueString(),
		Name:             data.Name.ValueString(),
		IsDeliveryDay:    data.IsDeliveryDay.ValueBool(),
		IsForAllZones:    data.IsForAllZones.ValueBool(),
		TimeZoneID:       data.TimeZoneID.ValueString(),
		DeliveryDayShift: int(data.DeliveryDayShift.ValueInt64()),
	}

	// Set zone_id if provided and not using is_for_all_zones
	if !data.ZoneID.IsNull() && !data.IsForAllZones.ValueBool() {
		deliveryTime.ZoneID = data.ZoneID.ValueString()
	}

	// Parse day if provided
	if !data.Day.IsNull() {
		var dayModel DeliveryDayModel
		diags.Append(data.Day.As(ctx, &dayModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil
		}

		// API requires ONLY ONE of: singleDate, datePeriod, or weekday
		// Priority: singleDate > datePeriod > weekday
		if !dayModel.Date.IsNull() && !dayModel.Date.IsUnknown() {
			// Specific date - use singleDate field
			deliveryTime.Day = &DeliveryDay{
				SingleDate: dayModel.Date.ValueString(),
			}
		} else if (!dayModel.DateFrom.IsNull() && !dayModel.DateFrom.IsUnknown()) || (!dayModel.DateTo.IsNull() && !dayModel.DateTo.IsUnknown()) {
			// Date range - use datePeriod object
			// Validate that both date_from and date_to are provided
			hasDateFrom := !dayModel.DateFrom.IsNull() && !dayModel.DateFrom.IsUnknown()
			hasDateTo := !dayModel.DateTo.IsNull() && !dayModel.DateTo.IsUnknown()

			if hasDateFrom && !hasDateTo {
				diags.AddError(
					"Invalid Date Range",
					"When using date_from, date_to must also be provided for a complete date range",
				)
				return nil
			}
			if !hasDateFrom && hasDateTo {
				diags.AddError(
					"Invalid Date Range",
					"When using date_to, date_from must also be provided for a complete date range",
				)
				return nil
			}

			// Both are provided, validate date_from <= date_to
			if hasDateFrom && hasDateTo {
				dateFrom := dayModel.DateFrom.ValueString()
				dateTo := dayModel.DateTo.ValueString()

				if dateFrom > dateTo {
					diags.AddError(
						"Invalid Date Range",
						fmt.Sprintf("date_from (%s) must be before or equal to date_to (%s)", dateFrom, dateTo),
					)
					return nil
				}

				deliveryTime.Day = &DeliveryDay{
					DatePeriod: &DatePeriod{
						DateFrom: dateFrom,
						DateTo:   dateTo,
					},
				}
			}
		} else if !dayModel.Weekday.IsNull() && !dayModel.Weekday.IsUnknown() {
			// Weekday only
			deliveryTime.Day = &DeliveryDay{
				Weekday: dayModel.Weekday.ValueString(),
			}
		}
	}

	// Parse slots if provided
	if !data.Slots.IsNull() {
		var slotsModels []DeliveryTimeSlotModel
		diags.Append(data.Slots.ElementsAs(ctx, &slotsModels, false)...)
		if diags.HasError() {
			return nil
		}

		for _, slotModel := range slotsModels {
			slot := DeliveryTimeSlot{
				ShippingMethod: slotModel.ShippingMethod.ValueString(),
				Capacity:       int(slotModel.Capacity.ValueInt64()),
			}

			// Parse delivery time range (REQUIRED field - must not be null/unknown)
			if slotModel.DeliveryTimeRange.IsNull() || slotModel.DeliveryTimeRange.IsUnknown() {
				diags.AddError(
					"Missing Required Field",
					"delivery_time_range is required for each slot",
				)
				return nil
			}

			var timeRangeModel TimeRangeModel
			diags.Append(slotModel.DeliveryTimeRange.As(ctx, &timeRangeModel, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil
			}

			// Validate time_from and time_to are not null
			if timeRangeModel.TimeFrom.IsNull() || timeRangeModel.TimeFrom.IsUnknown() {
				diags.AddError(
					"Missing Required Field",
					"time_from is required in delivery_time_range",
				)
				return nil
			}
			if timeRangeModel.TimeTo.IsNull() || timeRangeModel.TimeTo.IsUnknown() {
				diags.AddError(
					"Missing Required Field",
					"time_to is required in delivery_time_range",
				)
				return nil
			}

			slot.DeliveryTimeRange = &TimeRange{
				TimeFrom: timeRangeModel.TimeFrom.ValueString(),
				TimeTo:   timeRangeModel.TimeTo.ValueString(),
			}

			// Parse cut off time if provided (optional, but all fields required when present)
			if !slotModel.CutOffTime.IsNull() && !slotModel.CutOffTime.IsUnknown() {
				var cutOffTimeModel CutOffTimeModel
				diags.Append(slotModel.CutOffTime.As(ctx, &cutOffTimeModel, basetypes.ObjectAsOptions{})...)
				if diags.HasError() {
					return nil
				}

				// When cut_off_time is provided, both fields are required
				if cutOffTimeModel.Time.IsNull() || cutOffTimeModel.Time.IsUnknown() {
					diags.AddError(
						"Missing Required Field",
						"time is required in cut_off_time when cut_off_time is provided",
					)
					return nil
				}
				if cutOffTimeModel.DeliveryCycleName.IsNull() || cutOffTimeModel.DeliveryCycleName.IsUnknown() {
					diags.AddError(
						"Missing Required Field",
						"delivery_cycle_name is required in cut_off_time when cut_off_time is provided",
					)
					return nil
				}

				slot.CutOffTime = &CutOffTime{
					Time:              cutOffTimeModel.Time.ValueString(),
					DeliveryCycleName: cutOffTimeModel.DeliveryCycleName.ValueString(),
				}
			}

			deliveryTime.Slots = append(deliveryTime.Slots, slot)
		}
	}

	return deliveryTime
}

func (r *DeliveryTimeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeliveryTimeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating delivery time", map[string]interface{}{
		"name":      data.Name.ValueString(),
		"site_code": data.SiteCode.ValueString(),
	})

	// Validate zone_id requirement
	if !data.IsForAllZones.ValueBool() && data.ZoneID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"zone_id is required when is_for_all_zones is false. "+
				"Either provide a zone_id or set is_for_all_zones to true.",
		)
		return
	}

	// Build API request from model
	deliveryTime := buildDeliveryTimeFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create via API
	createdDeliveryTime, err := r.client.CreateDeliveryTime(ctx, deliveryTime)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create delivery time, got error: %s", err))
		return
	}

	// Ensure we got a valid response with ID
	if createdDeliveryTime == nil || createdDeliveryTime.ID == "" {
		resp.Diagnostics.AddError(
			"API Error",
			"API did not return an ID for the created delivery time. "+
				"Cannot proceed without a unique identifier.",
		)
		return
	}

	// Store the generated ID
	data.ID = types.StringValue(createdDeliveryTime.ID)

	// Read back using the ID to get actual state
	tflog.Debug(ctx, "Reading back created delivery time using ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
	actualDeliveryTime, err := r.client.GetDeliveryTime(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created delivery time, got error: %s", err))
		return
	}

	// Create fresh state model from API response (don't reuse plan data)
	var stateModel DeliveryTimeResourceModel
	r.syncModelFromAPI(ctx, &stateModel, actualDeliveryTime, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateModel)...)
}

func (r *DeliveryTimeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeliveryTimeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading delivery time", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	actualDeliveryTime, err := r.client.GetDeliveryTime(ctx, data.ID.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read delivery time, got error: %s", err))
		return
	}

	// Create fresh state model from API response
	var stateModel DeliveryTimeResourceModel
	r.syncModelFromAPI(ctx, &stateModel, actualDeliveryTime, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateModel)...)
}

func (r *DeliveryTimeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeliveryTimeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating delivery time", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	// Validate zone_id requirement
	if !data.IsForAllZones.ValueBool() && data.ZoneID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"zone_id is required when is_for_all_zones is false. "+
				"Either provide a zone_id or set is_for_all_zones to true.",
		)
		return
	}

	// Build update request from model
	deliveryTime := buildDeliveryTimeFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update via API
	_, err := r.client.UpdateDeliveryTime(ctx, data.ID.ValueString(), deliveryTime)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update delivery time, got error: %s", err))
		return
	}

	// Read back using ID
	tflog.Debug(ctx, "Reading back updated delivery time using ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
	actualDeliveryTime, err := r.client.GetDeliveryTime(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated delivery time, got error: %s", err))
		return
	}

	// Create fresh state model from API response (don't reuse plan data)
	var stateModel DeliveryTimeResourceModel
	r.syncModelFromAPI(ctx, &stateModel, actualDeliveryTime, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateModel)...)
}

func (r *DeliveryTimeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeliveryTimeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting delivery time", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	err := r.client.DeleteDeliveryTime(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete delivery time, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "Successfully deleted delivery time")
}

func (r *DeliveryTimeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: "id"
	// Example: terraform import emporix_delivery_time.example abc123
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// syncModelFromAPI syncs the Terraform model from API response
func (r *DeliveryTimeResource) syncModelFromAPI(ctx context.Context, model *DeliveryTimeResourceModel, api *DeliveryTime, diags *diag.Diagnostics) {
	model.ID = types.StringValue(api.ID)
	model.Name = types.StringValue(api.Name)
	model.SiteCode = types.StringValue(api.SiteCode)
	model.IsDeliveryDay = types.BoolValue(api.IsDeliveryDay)
	model.IsForAllZones = types.BoolValue(api.IsForAllZones)
	model.TimeZoneID = types.StringValue(api.TimeZoneID)
	model.DeliveryDayShift = types.Int64Value(int64(api.DeliveryDayShift))

	if api.ZoneID != "" {
		model.ZoneID = types.StringValue(api.ZoneID)
	} else {
		model.ZoneID = types.StringNull()
	}

	// Sync day
	if api.Day != nil {
		dayModel := DeliveryDayModel{}

		// Set weekday if present
		if api.Day.Weekday != "" {
			dayModel.Weekday = types.StringValue(api.Day.Weekday)
		} else {
			dayModel.Weekday = types.StringNull()
		}

		// Set single date if present
		if api.Day.SingleDate != "" {
			dayModel.Date = types.StringValue(api.Day.SingleDate)
		} else {
			dayModel.Date = types.StringNull()
		}

		// Set date period if present
		if api.Day.DatePeriod != nil {
			if api.Day.DatePeriod.DateFrom != "" {
				dayModel.DateFrom = types.StringValue(api.Day.DatePeriod.DateFrom)
			} else {
				dayModel.DateFrom = types.StringNull()
			}

			if api.Day.DatePeriod.DateTo != "" {
				dayModel.DateTo = types.StringValue(api.Day.DatePeriod.DateTo)
			} else {
				dayModel.DateTo = types.StringNull()
			}
		} else {
			dayModel.DateFrom = types.StringNull()
			dayModel.DateTo = types.StringNull()
		}

		dayObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"weekday":   types.StringType,
			"date":      types.StringType,
			"date_from": types.StringType,
			"date_to":   types.StringType,
		}, dayModel)
		diags.Append(d...)
		model.Day = dayObj
	} else {
		model.Day = types.ObjectNull(map[string]attr.Type{
			"weekday":   types.StringType,
			"date":      types.StringType,
			"date_from": types.StringType,
			"date_to":   types.StringType,
		})
	}

	// Sync slots
	if len(api.Slots) > 0 {
		var slotsModels []DeliveryTimeSlotModel
		for _, slot := range api.Slots {
			slotModel := DeliveryTimeSlotModel{
				ShippingMethod: types.StringValue(slot.ShippingMethod),
				Capacity:       types.Int64Value(int64(slot.Capacity)),
			}

			// Time range
			if slot.DeliveryTimeRange != nil {
				timeRangeModel := TimeRangeModel{
					TimeFrom: types.StringValue(slot.DeliveryTimeRange.TimeFrom),
					TimeTo:   types.StringValue(slot.DeliveryTimeRange.TimeTo),
				}
				timeRangeObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
					"time_from": types.StringType,
					"time_to":   types.StringType,
				}, timeRangeModel)
				diags.Append(d...)
				slotModel.DeliveryTimeRange = timeRangeObj
			} else {
				// If API returns nil, set to ObjectNull (though this shouldn't happen for required field)
				slotModel.DeliveryTimeRange = types.ObjectNull(map[string]attr.Type{
					"time_from": types.StringType,
					"time_to":   types.StringType,
				})
			}

			// Cut off time
			if slot.CutOffTime != nil {
				cutOffTimeModel := CutOffTimeModel{
					Time:              types.StringValue(slot.CutOffTime.Time),
					DeliveryCycleName: types.StringValue(slot.CutOffTime.DeliveryCycleName),
				}
				cutOffTimeObj, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
					"time":                types.StringType,
					"delivery_cycle_name": types.StringType,
				}, cutOffTimeModel)
				diags.Append(d...)
				slotModel.CutOffTime = cutOffTimeObj
			} else {
				slotModel.CutOffTime = types.ObjectNull(map[string]attr.Type{
					"time":                types.StringType,
					"delivery_cycle_name": types.StringType,
				})
			}

			slotsModels = append(slotsModels, slotModel)
		}

		slotsList, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"shipping_method": types.StringType,
				"capacity":        types.Int64Type,
				"delivery_time_range": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"time_from": types.StringType,
						"time_to":   types.StringType,
					},
				},
				"cut_off_time": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"time":                types.StringType,
						"delivery_cycle_name": types.StringType,
					},
				},
			},
		}, slotsModels)
		diags.Append(d...)
		model.Slots = slotsList
	} else {
		// Use empty list instead of null for consistency
		emptySlotsList, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"shipping_method": types.StringType,
				"capacity":        types.Int64Type,
				"delivery_time_range": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"time_from": types.StringType,
						"time_to":   types.StringType,
					},
				},
				"cut_off_time": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"time":                types.StringType,
						"delivery_cycle_name": types.StringType,
					},
				},
			},
		}, []DeliveryTimeSlotModel{})
		diags.Append(d...)
		model.Slots = emptySlotsList
	}
}
