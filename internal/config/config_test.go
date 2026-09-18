package config

import "testing"

func Test_Load_requires_api_key(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "")
	t.Setenv("JEV_MCP_MODEL", "")
	t.Setenv("TYPESAFE_BASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}

func Test_Load_uses_defaults_when_only_key_is_set(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "ts_live")
	t.Setenv("JEV_MCP_MODEL", "")
	t.Setenv("TYPESAFE_BASE_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "ts_live" || cfg.Model != DefaultModel || cfg.BaseURL != DefaultBaseURL {
		t.Fatalf("%+v", cfg)
	}
}

func Test_Load_honors_model_and_base_url(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "ts_live")
	t.Setenv("JEV_MCP_MODEL", "jev-1.13.0")
	t.Setenv("TYPESAFE_BASE_URL", "https://example.test/")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "jev-1.13.0" || cfg.BaseURL != "https://example.test" {
		t.Fatalf("%+v", cfg)
	}
}
