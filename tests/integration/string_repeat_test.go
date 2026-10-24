package integration_test

import (
	"strings"
	"testing"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

func TestStringRepeat(t *testing.T) {
	const limit = 100 * 1024 * 1024
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name    string
		value   string
		count   int
		want    string
		wantErr bool
	}{
		{name: "ordinary", value: "ab", count: 3, want: "ababab"},
		{name: "one", value: "ab", count: 1, want: "ab"},
		{name: "zero", value: "ab", count: 0},
		{name: "negative", value: "ab", count: -1},
		{name: "minimum_negative", value: "ab", count: -maxInt - 1},
		{name: "empty", count: 3},
		{name: "empty_zero", count: 0},
		{name: "empty_negative", count: -1},
		{name: "empty_maximum_count", count: maxInt},
		{name: "multibyte", value: "\u00e9", count: 3, want: "\u00e9\u00e9\u00e9"},
		{name: "above_limit", value: "A", count: limit + 1, wantErr: true},
		{name: "multibyte_above_limit", value: "\u00e9", count: limit/2 + 1, wantErr: true},
		{name: "multibyte_rounding", value: "\u20ac", count: limit/3 + 1, wantErr: true},
		{name: "overflow", value: "ABCD", count: maxInt/2 + 1, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if value := recover(); value != nil {
					t.Fatalf("unexpected panic: %v", value)
				}
			}()
			source := `{{ value * count }}`
			if tc.wantErr {
				// Do not render a large result if the protection regresses.
				source = `{% set result = value * count %}`
			}
			template, err := gonja.FromString(source)
			if err != nil {
				t.Fatal(err)
			}
			got, err := template.ExecuteToString(exec.NewContext(map[string]any{
				"value": tc.value,
				"count": tc.count,
			}))
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "string repeat exceeds limit of 104857600 bytes") {
					t.Fatalf("expected string repeat size-limit error, got %v", err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got (%q, %v), want (%q, nil)", got, err, tc.want)
			}
		})
	}
}

func TestStringRepeatArithmetic(t *testing.T) {
	for _, tc := range []struct{ name, source, want string }{
		{"integer", `{{ 6 * 7 }}`, "42"},
		{"negative_integer", `{{ 6 * -7 }}`, "-42"},
		{"large_integer", `{{ 1000000000 * 2 }}`, "2000000000"},
		{"float_left", `{{ 2.5 * 3 }}`, "7.5"},
		{"float_right", `{{ 3 * 2.5 }}`, "7.5"},
		{"precedence", `{{ 3 + 4 * 2 }}`, "11"},
		{"subtraction", `{{ 10 - 3 }}`, "7"},
		{"division", `{{ 7 / 2 }}`, "3.5"},
		{"floor_division", `{{ 7 // 2 }}`, "3"},
		{"modulo", `{{ 7 % 3 }}`, "1"},
		{"power", `{{ 2 ** 3 }}`, "8.0"},
		{"concatenation", `{{ 'a' + 'b' }}`, "ab"},
		{"nested_repeat", `{{ ('ab' * 2) * 2 }}`, "abababab"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			template, err := gonja.FromString(tc.source)
			if err != nil {
				t.Fatal(err)
			}
			got, err := template.ExecuteToString(nil)
			if err != nil || got != tc.want {
				t.Fatalf("got (%q, %v), want (%q, nil)", got, err, tc.want)
			}
		})
	}
}
