package runner

import (
	"reflect"
	"testing"

	"github.com/rliebz/tusk/marshal"
	yaml "gopkg.in/yaml.v2"
)

func TestOption_Dependencies(t *testing.T) {
	option := &Option{DefaultValues: ValueList{
		{When: WhenList{whenFalse}, Value: "foo"},
		{When: WhenList{createWhen(
			withWhenEqual("foo", "foovalue"),
			withWhenEqual("bar", "barvalue"),
		)}, Value: "bar"},
		{When: WhenList{createWhen(
			withWhenNotEqual("baz", "bazvalue"),
		)}, Value: "bar"},
	}}

	expected := []string{"foo", "bar", "baz"}
	actual := option.Dependencies()
	if !equalUnordered(expected, actual) {
		t.Errorf(
			"Option.Dependencies(): expected %s, actual %s",
			expected, actual,
		)
	}
}

func equalUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]interface{})
	for _, val := range a {
		aMap[val] = struct{}{}
	}

	bMap := make(map[string]interface{})
	for _, val := range b {
		bMap[val] = struct{}{}
	}

	return reflect.DeepEqual(aMap, bMap)
}

// Env var `OPTION_VAR` will be set to `option_val`
var valuetests = []struct {
	desc     string
	input    *Option
	expected string
}{
	{"nil", nil, ""},
	{"empty option", &Option{}, ""},
	{
		"default only",
		&Option{DefaultValues: ValueList{
			{Value: "default"},
		}},
		"default",
	},
	{
		"command only",
		&Option{DefaultValues: ValueList{
			{Command: "echo command"},
		}},
		"command",
	},
	{
		"environment variable only",
		&Option{Environment: "OPTION_VAR"},
		"option_val",
	},
	{
		"passed variable only",
		&Option{Passable: Passable{Passed: "passed"}},
		"passed",
	},
	{
		"conditional value",
		&Option{DefaultValues: ValueList{
			{When: WhenList{whenFalse}, Value: "foo"},
			{When: WhenList{whenTrue}, Value: "bar"},
			{When: WhenList{whenFalse}, Value: "baz"},
		}},
		"bar",
	},
	{
		"passed when all settings are defined",
		&Option{
			Environment: "OPTION_VAR",
			DefaultValues: ValueList{
				{When: WhenList{whenTrue}, Value: "when"},
			},
			Passable: Passable{
				Passed: "passed",
			},
		},
		"passed",
	},
}

func TestOption_Evaluate(t *testing.T) {
	t.Setenv("OPTION_VAR", "option_val")

	for _, tt := range valuetests {
		actual, err := tt.input.Evaluate(Context{}, nil)
		if err != nil {
			t.Errorf(
				"Option.Evaluate() for %s: unexpected err: %q",
				tt.desc, err,
			)
			continue
		}

		if tt.expected != actual {
			t.Errorf(
				"Option.Evaluate() for %s: expected %q, actual %q",
				tt.desc, tt.expected, actual,
			)
		}
	}
}

func TestOption_Evaluate_required_nothing_passed(t *testing.T) {
	option := Option{Required: true}

	if _, err := option.Evaluate(Context{}, nil); err == nil {
		t.Fatal(
			"Option.Evaluate() for required option: expected err, actual nil",
		)
	}
}

func TestOption_Evaluate_passes_vars(t *testing.T) {
	expected := "some value"
	opt := Option{
		DefaultValues: ValueList{
			{When: WhenList{whenFalse}, Value: "wrong"},
			{
				When:  WhenList{createWhen(withWhenEqual("foo", "foovalue"))},
				Value: expected,
			},
			{When: WhenList{whenFalse}, Value: "oops"},
		},
	}

	actual, err := opt.Evaluate(Context{}, map[string]string{"foo": "foovalue"})
	if err != nil {
		t.Fatalf("Option.Evaluate(): unexpected error: %s", err)
	}

	if expected != actual {
		t.Errorf(
			"Option.Evaluate(): expected %q, actual %q",
			expected, actual,
		)
	}
}

func TestOption_Evaluate_required_with_passed(t *testing.T) {
	expected := "foo"
	option := Option{
		Required: true,
		Passable: Passable{
			Passed: expected,
		},
	}

	actual, err := option.Evaluate(Context{}, nil)
	if err != nil {
		t.Fatalf("Option.Evaluate(): unexpected error: %s", err)
	}

	if expected != actual {
		t.Errorf(
			"Option.Evaluate(): expected %q, actual %q",
			expected, actual,
		)
	}
}

func TestOption_Evaluate_required_with_environment(t *testing.T) {
	envVar := "OPTION_VAR"
	expected := "foo"

	option := Option{Required: true, Environment: envVar}
	t.Setenv(envVar, expected)

	actual, err := option.Evaluate(Context{}, nil)
	if err != nil {
		t.Fatalf("Option.Evaluate(): unexpected error: %s", err)
	}

	if expected != actual {
		t.Errorf(
			"Option.Evaluate(): expected %q, actual %q",
			expected, actual,
		)
	}
}

func TestOption_Evaluate_values_none_specified(t *testing.T) {
	expected := ""
	option := Option{
		Passable: Passable{
			ValuesAllowed: marshal.StringList{"red", "herring"},
		},
	}

	actual, err := option.Evaluate(Context{}, nil)
	if err != nil {
		t.Fatalf("Option.Evaluate(): unexpected error: %s", err)
	}

	if expected != actual {
		t.Errorf(
			"Option.Evaluate(): expected %q, actual %q",
			expected, actual,
		)
	}
}

func TestOption_Evaluate_values_with_passed(t *testing.T) {
	expected := "foo"
	option := Option{
		Passable: Passable{
			Passed:        expected,
			ValuesAllowed: marshal.StringList{"red", expected, "herring"},
		},
	}

	actual, err := option.Evaluate(Context{}, nil)
	if err != nil {
		t.Fatalf("Option.Evaluate(): unexpected error: %s", err)
	}

	if expected != actual {
		t.Errorf(
			"Option.Evaluate(): expected %q, actual %q",
			expected, actual,
		)
	}
}

func TestOption_Evaluate_values_with_environment(t *testing.T) {
	envVar := "OPTION_VAR"
	expected := "foo"

	option := Option{
		Environment: envVar,
		Passable: Passable{
			ValuesAllowed: marshal.StringList{"red", expected, "herring"},
		},
	}

	t.Setenv(envVar, expected)

	actual, err := option.Evaluate(Context{}, nil)
	if err != nil {
		t.Fatalf("Option.Evaluate(): unexpected error: %s", err)
	}

	if expected != actual {
		t.Errorf(
			"Option.Evaluate(): expected %q, actual %q",
			expected, actual,
		)
	}
}

func TestOption_Evaluate_values_with_invalid_passed(t *testing.T) {
	expected := "foo"
	option := Option{
		Passable: Passable{
			Passed:        expected,
			ValuesAllowed: marshal.StringList{"bad", "values", "FOO"},
		},
	}

	_, err := option.Evaluate(Context{}, nil)
	if err == nil {
		t.Fatalf(
			"Option.Evaluate(): expected error for invalid passed value, got nil",
		)
	}
}

func TestOption_Evaluate_values_with_invalid_environment(t *testing.T) {
	envVar := "OPTION_VAR"
	expected := "foo"

	option := Option{
		Environment: envVar,
		Passable: Passable{
			ValuesAllowed: marshal.StringList{"bad", "values", "FOO"},
		},
	}

	t.Setenv(envVar, expected)

	_, err := option.Evaluate(Context{}, nil)
	if err == nil {
		t.Fatalf(
			"Option.Evaluate(): expected error for invalid environment value, got nil",
		)
	}
}

var evaluteTypeDefaultTests = []struct {
	typeName string
	expected string
}{
	{"int", "0"},
	{"INTEGER", "0"},
	{"Float", "0"},
	{"float64", "0"},
	{"double", "0"},
	{"bool", "false"},
	{"boolean", "false"},
	{"", ""},
}

func TestOption_Evaluate_type_defaults(t *testing.T) {
	for _, tt := range evaluteTypeDefaultTests {
		opt := Option{
			Passable: Passable{
				Type: tt.typeName,
			},
		}
		actual, err := opt.Evaluate(Context{}, nil)
		if err != nil {
			t.Errorf("Option.Evaluate(): unexpected error: %s", err)
			continue
		}

		if tt.expected != actual {
			t.Errorf(
				"Option.Evaluate(): expected %q, actual %q",
				tt.expected, actual,
			)
		}
	}
}

func TestOption_UnmarshalYAML(t *testing.T) {
	s := []byte(`{usage: foo, values: [foo, bar]}`)
	expected := Option{
		Passable: Passable{
			Name:          "",
			Usage:         "foo",
			ValuesAllowed: []string{"foo", "bar"},
		},
	}
	actual := Option{}

	if err := yaml.UnmarshalStrict(s, &actual); err != nil {
		t.Fatalf("yaml.UnmarshalStrict(%s, ...): unexpected error: %s", s, err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(
			`yaml.UnmarshalStrict(%s, ...): expected "%#v", actual "%#v"`,
			s, expected, actual,
		)
	}
}

var unmarshalOptionErrorTests = []struct {
	desc  string
	input string
}{
	{
		"invalid option definition",
		"string only",
	},
	{
		"short name exceeds one character",
		"{short: foo}",
	},
	{
		"private and required defined",
		"{private: true, required: true}",
	},
	{
		"private and environment defined",
		"{private: true, environment: ENV_VAR}",
	},
	{
		"private and values defined",
		"{private: true, values: [foo, bar]}",
	},
	{
		"required and default defined",
		"{required: true, default: foo}",
	},
}

func TestOption_UnmarshalYAML_invalid_definitions(t *testing.T) {
	for _, tt := range unmarshalOptionErrorTests {
		o := Option{}
		if err := yaml.UnmarshalStrict([]byte(tt.input), &o); err == nil {
			t.Errorf(
				"yaml.UnmarshalStrict(%s, ...): expected error for %s, actual nil",
				tt.input, tt.desc,
			)
		}
	}
}

func TestGetOptionsWithOrder(t *testing.T) {
	name := "foo"
	env := "fooenv"
	ms := yaml.MapSlice{
		{Key: name, Value: &Option{Environment: env}},
		{Key: "bar", Value: &Option{Environment: "barenv"}},
	}

	options, err := getOptionsWithOrder(ms)
	if err != nil {
		t.Fatalf("GetOptionsWithOrder(ms) => unexpected error: %v", err)
	}

	if len(ms) != len(options) {
		t.Fatalf(
			"GetOptionsWithOrder(ms) => want %d items, got %d",
			len(ms), len(options),
		)
	}

	opt := options[0]

	if name != opt.Name {
		t.Errorf(
			"GetOptionsWithOrder(ms) => want opt.Name %q, got %q",
			name, opt.Name,
		)
	}

	if env != opt.Environment {
		t.Errorf(
			"GetOptionsWithOrder(ms) => want opt.Environment %q, got %q",
			env, opt.Environment,
		)
	}

	if options[1].Name != "bar" {
		t.Errorf("GetOptionsWithOrder(ms) => want 2nd option %q, got %q", "bar", options[1].Name)
	}
}
