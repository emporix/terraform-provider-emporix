package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var mixinObjectAttrTypes = map[string]attr.Type{
	"name":       types.StringType,
	"schema_url": types.StringType,
	"fields":     types.StringType,
}

func mixinsListValue(t *testing.T, names []string) types.List {
	t.Helper()

	models := make([]MixinModel, 0, len(names))
	for _, n := range names {
		models = append(models, MixinModel{
			Name:      types.StringValue(n),
			SchemaURL: types.StringValue("https://example.com/" + n),
			Fields:    types.StringValue(`{}`),
		})
	}

	list, d := types.ListValueFrom(context.Background(), types.ObjectType{AttrTypes: mixinObjectAttrTypes}, models)
	if d.HasError() {
		t.Fatalf("failed to build mixins list: %v", d)
	}
	return list
}

// mixinNamesFromList extracts mixin names, in list order, from a types.List of MixinModel.
func mixinNamesFromList(t *testing.T, list types.List) []string {
	t.Helper()

	var mixins []MixinModel
	if d := list.ElementsAs(context.Background(), &mixins, false); d.HasError() {
		t.Fatalf("failed to read mixins list: %v", d)
	}

	names := make([]string, 0, len(mixins))
	for _, m := range mixins {
		names = append(names, m.Name.ValueString())
	}
	return names
}

// TestOrderedMixinNames_PreservesPreviousOrder guards against the "Provider
// produced inconsistent result after apply" bug: with 2+ mixins, previously
// the list order was derived from iterating a Go map directly, which is
// randomized and could differ between calls even for identical input.
func TestOrderedMixinNames_PreservesPreviousOrder(t *testing.T) {
	ctx := context.Background()

	metadataMixins := map[string]string{
		"site-custom-fields1": "https://schema/1",
		"site-custom-fields2": "https://schema/2",
		"site-custom-fields3": "https://schema/3",
		"site-custom-fields4": "https://schema/4",
		"site-custom-fields5": "https://schema/5",
	}
	previousOrder := []string{
		"site-custom-fields3",
		"site-custom-fields1",
		"site-custom-fields5",
		"site-custom-fields2",
		"site-custom-fields4",
	}
	previousModel := &SiteSettingsResourceModel{
		Mixins: mixinsListValue(t, previousOrder),
	}

	for i := 0; i < 50; i++ {
		got := orderedMixinNames(ctx, metadataMixins, previousModel)
		if !reflect.DeepEqual(got, previousOrder) {
			t.Fatalf("iteration %d: orderedMixinNames = %v, want %v", i, got, previousOrder)
		}
	}
}

func TestOrderedMixinNames_NewMixinsAppendedAlphabetically(t *testing.T) {
	ctx := context.Background()

	metadataMixins := map[string]string{
		"zeta":  "https://schema/zeta",
		"alpha": "https://schema/alpha",
		"beta":  "https://schema/beta",
	}
	previousModel := &SiteSettingsResourceModel{
		Mixins: mixinsListValue(t, []string{"beta"}),
	}

	want := []string{"beta", "alpha", "zeta"}
	got := orderedMixinNames(ctx, metadataMixins, previousModel)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("orderedMixinNames = %v, want %v", got, want)
	}
}

func TestOrderedMixinNames_NoPreviousMixins_SortsAlphabetically(t *testing.T) {
	ctx := context.Background()

	metadataMixins := map[string]string{
		"c": "https://schema/c",
		"a": "https://schema/a",
		"b": "https://schema/b",
	}
	previousModel := &SiteSettingsResourceModel{
		Mixins: types.ListNull(types.ObjectType{AttrTypes: mixinObjectAttrTypes}),
	}

	want := []string{"a", "b", "c"}
	got := orderedMixinNames(ctx, metadataMixins, previousModel)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("orderedMixinNames = %v, want %v", got, want)
	}
}

// TestApiToTerraform_MixinsOrderStableAcrossRepeatedCalls reproduces the
// original bug report: with more than one mixin configured, repeated
// Create/Update runs against identical API data must always yield the mixins
// list in the same order, otherwise Terraform reports "Provider produced
// inconsistent result after apply".
func TestApiToTerraform_MixinsOrderStableAcrossRepeatedCalls(t *testing.T) {
	ctx := context.Background()
	names := []string{
		"site-custom-fields1",
		"site-custom-fields2",
		"site-custom-fields3",
		"site-custom-fields4",
		"site-custom-fields5",
	}

	site := &SiteSettings{
		Name:            "test",
		DefaultLanguage: "en",
		Currency:        "EUR",
		Metadata:        &Metadata{Mixins: map[string]string{}},
		Mixins:          map[string]interface{}{},
	}
	for _, n := range names {
		site.Metadata.Mixins[n] = "https://schema/" + n
		site.Mixins[n] = map[string]interface{}{"field": n}
	}

	r := &SiteSettingsResource{}

	for i := 0; i < 50; i++ {
		// Mirrors Create/Update: plan is passed as both model and previousModel.
		plan := &SiteSettingsResourceModel{
			Mixins: mixinsListValue(t, names),
		}
		var diags diag.Diagnostics

		r.apiToTerraform(ctx, site, plan, plan, &diags, true)
		if diags.HasError() {
			t.Fatalf("iteration %d: apiToTerraform returned errors: %v", i, diags)
		}

		got := mixinNamesFromList(t, plan.Mixins)
		if !reflect.DeepEqual(got, names) {
			t.Fatalf("iteration %d: mixins order = %v, want %v", i, got, names)
		}
	}
}