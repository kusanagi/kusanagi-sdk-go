package sdk

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kusanagi/kusanagi-sdk-go/logging"
	"github.com/kusanagi/kusanagi-sdk-go/transform"
)

type TestComponent struct {
	res bool
}

func (c *TestComponent) HasResource(name string) bool                 { return c.res }
func (c *TestComponent) GetResource(name string) (interface{}, error) { return c.res, nil }

func TestAPI(t *testing.T) {
	path := "/path/to/file.go"
	name := "dummy"
	version := "1.0"
	frameworkVersion := "1.0.0"

	c := &TestComponent{}
	api := newApi(c, path, name, version, frameworkVersion)
	api.debug = true
	if !api.IsDebug() {
		t.Error("debug should be true")
	}
	if v := api.GetFrameworkVersion(); v != frameworkVersion {
		t.Errorf("invalid value, expected: %s, got: %s", frameworkVersion, v)
	}
	if v := api.GetPath(); v != path {
		t.Errorf("invalid value, expected: %s, got: %s", path, v)
	}
	if v := api.GetName(); v != name {
		t.Errorf("invalid value, expected: %s, got: %s", name, v)
	}
	if v := api.GetVersion(); v != version {
		t.Errorf("invalid value, expected: %s, got: %s", version, v)
	}

	if api.GetVariables() != nil {
		t.Error("variables should be empty")
	}
	if api.HasVariable("foo") {
		t.Error("the variable 'foo' should't exist")
	}
	if api.GetVariable("foo") != "" {
		t.Error("the variable 'foo' should be empty")
	}
	api.variables = Variables{"foo": "bar"}
	if api.GetVariables() == nil {
		t.Error("variables should not be empty")
	}
	if !api.HasVariable("foo") {
		t.Error("the variable 'foo' should exist")
	}
	if v := api.GetVariable("foo"); v != "bar" {
		t.Errorf("invalid 'foo' variable value, got: %s, expected: bar", v)
	}

	api.component = &TestComponent{true}
	if !api.HasResource("foo") {
		t.Error("resource should exist")
	}
	if v, _ := api.GetResource("foo").(bool); !v {
		t.Error("invalid resource value, expedted true")
	}
}

func TestAPIGetServices(t *testing.T) {
	path := "/path/to/file.go"
	name := "dummy"
	version := "1.0"
	frameworkVersion := "1.0.0"

	api := newApi(&TestComponent{}, path, name, version, frameworkVersion)
	api.debug = true
	services := api.GetServices()
	if len(services) != 0 {
		t.Error("services should be empty")
	}

	// Add mappings to the schema registry
	data, err := transform.Pack(map[string]interface{}{
		"foo": map[string]interface{}{
			"1.0.0": make(map[string]interface{}),
		},
	})
	if err != nil {
		t.Fatalf("failed to serialize the schema registry data: %v", err)
	}
	api.registry.Update(data)
	services = api.GetServices()
	if v := len(services); v != 1 {
		t.Fatalf("only a service was expected, got: %d", v)
	}

	// Check that the service info is right
	service := services[0]
	if v, _ := service["name"]; v != "foo" {
		t.Errorf("invalid service name, expected: foo, got: %s", v)
	}
	if v, _ := service["version"]; v != "1.0.0" {
		t.Errorf("invalid service name, expected: 1.0.0, got: %s", v)
	}
}

func TestAPIGetServiceSchema(t *testing.T) {
	t.Skip("TODO")
}

func TestAPILog(t *testing.T) {
	path := "/path/to/file.go"
	name := "dummy"
	version := "1.0"
	frameworkVersion := "1.0.0"

	buf := bytes.Buffer{}
	logging.SetOutput(&buf)
	defer logging.Disable()
	logging.SetLevel(logging.DEBUG)
	defer logging.SetLevel(logging.EMERGENCY)

	api := newApi(&TestComponent{}, path, name, version, frameworkVersion)
	msg := "Test message"
	api.Log(msg, logging.INFO)
	if v := buf.String(); v != "" {
		t.Fatalf("log should be empty, got: %s", v)
	}

	// When debug is true log should appear
	api.debug = true
	api.Log(msg, logging.INFO)
	if v := buf.String(); !strings.HasSuffix(strings.TrimRight(v, "\n"), msg) {
		t.Fatalf("log does not end with the logged message, got: %s", v)
	}
}
