package provider

import (
	"context"
	"testing"
	"time"
)

func TestDurationSemanticEquals(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"5m", "5m0s", true},   // canonical vs shorthand
		{"5m", "300s", true},   // minutes vs seconds
		{"1h", "60m", true},    // hours vs minutes
		{"5m", "6m", false},    // genuinely different
		{"1h30m", "90m", true}, // compound
	}
	for _, tc := range cases {
		eq, diags := newDurationValue(tc.a).StringSemanticEquals(context.Background(), newDurationValue(tc.b))
		if diags.HasError() {
			t.Fatalf("%s vs %s: diags %v", tc.a, tc.b, diags)
		}
		if eq != tc.want {
			t.Errorf("%s == %s: got %v, want %v", tc.a, tc.b, eq, tc.want)
		}
	}
}

func TestSecondsToDuration(t *testing.T) {
	if v := secondsToDuration(0); !v.IsNull() {
		t.Errorf("0 seconds should map to null, got %v", v)
	}
	v := secondsToDuration(300)
	if v.IsNull() || v.ValueString() != "5m0s" {
		t.Errorf("300s => %q; want 5m0s", v.ValueString())
	}
	// And it must be semantically equal to a user's "5m".
	eq, _ := v.StringSemanticEquals(context.Background(), newDurationValue("5m"))
	if !eq {
		t.Error("secondsToDuration(300) should equal \"5m\"")
	}
}

func TestDurationParse(t *testing.T) {
	// Mirrors the type's Validate path (time.ParseDuration).
	if _, err := time.ParseDuration("nope"); err == nil {
		t.Error("expected parse error for \"nope\"")
	}
	if _, err := time.ParseDuration("90m"); err != nil {
		t.Errorf("90m should parse: %v", err)
	}
}
