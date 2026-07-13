package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// durationType is a string type whose values compare by parsed time.Duration.
// The API accepts Go duration strings ("5m") on write but returns seconds on
// read, which the resource renders back as a canonical string ("5m0s"). Without
// semantic equality that round-trip would show a perpetual diff; with it,
// "5m" == "5m0s" == "300s" and the plan stays clean.
type durationType struct {
	basetypes.StringType
}

var (
	_ basetypes.StringTypable = durationType{}
	_ xattr.TypeWithValidate  = durationType{}
)

func (t durationType) String() string { return "provider.durationType" }

func (t durationType) ValueType(context.Context) attr.Value {
	return durationValue{}
}

func (t durationType) Equal(o attr.Type) bool {
	_, ok := o.(durationType)
	return ok
}

func (t durationType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return durationValue{StringValue: in}, nil
}

func (t durationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	sv, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", attrValue)
	}
	return durationValue{StringValue: sv}, nil
}

// Validate rejects a non-empty value that time.ParseDuration can't parse, so
// bad input fails at plan time with a clear message instead of a later API 400.
func (t durationType) Validate(_ context.Context, in tftypes.Value, valPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if !in.IsKnown() || in.IsNull() {
		return diags
	}
	var s string
	if err := in.As(&s); err != nil {
		diags.AddAttributeError(valPath, "Invalid Duration", err.Error())
		return diags
	}
	if _, err := time.ParseDuration(s); err != nil {
		diags.AddAttributeError(valPath, "Invalid Duration",
			fmt.Sprintf("%q is not a valid Go duration (e.g. \"30s\", \"5m\", \"1h30m\"): %s", s, err))
	}
	return diags
}

// durationValue carries the semantic-equality behavior.
type durationValue struct {
	basetypes.StringValue
}

var (
	_ basetypes.StringValuable                   = durationValue{}
	_ basetypes.StringValuableWithSemanticEquals = durationValue{}
)

func (v durationValue) Type(context.Context) attr.Type { return durationType{} }

func (v durationValue) Equal(o attr.Value) bool {
	other, ok := o.(durationValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals treats two durations as equal when they parse to the
// same time.Duration, regardless of spelling.
func (v durationValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	newVal, ok := newValuable.(durationValue)
	if !ok {
		return false, diags
	}
	if v.IsNull() || v.IsUnknown() || newVal.IsNull() || newVal.IsUnknown() {
		return v.StringValue.Equal(newVal.StringValue), diags
	}
	a, err := time.ParseDuration(v.ValueString())
	if err != nil {
		return false, diags
	}
	b, err := time.ParseDuration(newVal.ValueString())
	if err != nil {
		return false, diags
	}
	return a == b, diags
}

// newDurationValue wraps a plain string as a durationValue.
func newDurationValue(s string) durationValue {
	return durationValue{StringValue: basetypes.NewStringValue(s)}
}

func newDurationNull() durationValue {
	return durationValue{StringValue: basetypes.NewStringNull()}
}

// secondsToDuration renders an API seconds count as a canonical duration
// string, or a null durationValue when zero (an unset optional field).
func secondsToDuration(secs int64) durationValue {
	if secs <= 0 {
		return newDurationNull()
	}
	return newDurationValue((time.Duration(secs) * time.Second).String())
}
