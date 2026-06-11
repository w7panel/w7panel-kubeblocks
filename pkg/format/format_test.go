package format

import "testing"

func TestFormatJSON(t *testing.T) {
	data := []byte(`{"port":3306,"tls":{"enabled":true},"hosts":["a","b"]}`)

	got, err := Format(data, JSON, "test")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := map[string]string{
		"port":        "3306",
		"tls.enabled": "true",
		"hosts[0]":    "a",
		"hosts[1]":    "b",
	}
	assertMapContains(t, got, want)
}

func TestFormatYAML(t *testing.T) {
	data := []byte("server:\n  port: 8080\n  tls: true\n")

	got, err := Format(data, YAML, "test")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := map[string]string{
		"server.port": "8080",
		"server.tls":  "true",
	}
	assertMapContains(t, got, want)
}

func TestFormatINI(t *testing.T) {
	data := []byte("port=3306\n[client]\nuser=root\n")

	got, err := Format(data, Ini, "test")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := map[string]string{
		"port":        "3306",
		"client.user": "root",
	}
	assertMapContains(t, got, want)
}

func TestFormatDotenv(t *testing.T) {
	data := []byte("USER=root\nPASSWORD=secret\n")

	got, err := Format(data, Dotenv, "test")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := map[string]string{
		"USER":     "root",
		"PASSWORD": "secret",
	}
	assertMapContains(t, got, want)
}

func TestFormatProperties(t *testing.T) {
	data := []byte("db.port=5432\ncamelCaseKey=value\n")

	got, err := Format(data, PropertiesPlus, "test")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := map[string]string{
		"db.port":      "5432",
		"camelCaseKey": "value",
	}
	assertMapContains(t, got, want)
}

func TestFormatTOML(t *testing.T) {
	data := []byte("[server]\nport = 7001\nenabled = true\n")

	got, err := Format(data, TOML, "test")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := map[string]string{
		"server.port":    "7001",
		"server.enabled": "true",
	}
	assertMapContains(t, got, want)
}

func TestFormatXML(t *testing.T) {
	data := []byte("<config><server port=\"8080\">mysql</server></config>")

	got, err := Format(data, XML, "test")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := map[string]string{
		"config.server.-port": "8080",
		"config.server.#text": "mysql",
	}
	assertMapContains(t, got, want)
}

func TestFormatRedis(t *testing.T) {
	data := []byte("bind 0.0.0.0\nappendonly yes\n# comment\n")

	got, err := Format(data, RedisCfg, "test")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := map[string]string{
		"bind":       "0.0.0.0",
		"appendonly": "yes",
	}
	assertMapContains(t, got, want)
}

func assertMapContains(t *testing.T, got, want map[string]string) {
	t.Helper()

	for key, wantValue := range want {
		if gotValue, ok := got[key]; !ok || gotValue != wantValue {
			t.Fatalf("key %q = %q, want %q", key, gotValue, wantValue)
		}
	}
}
