package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ==============================
// Session tests (session.go)
// ==============================

func TestSaveAndLoadLastRequest(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "hsp-test-save-load")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)
	os.Setenv("HOME", testDir)

	rb := &RequestBuilder{
		URL:         "https://api.example.com/users",
		Method:      "POST",
		Headers:     map[string]string{"Authorization": "Bearer token123", "Content-Type": "application/json"},
		QueryParams: map[string]string{"include": "profile"},
		Body:        `{"name":"John"}`,
		BodyFormat:  "json",
	}

	if err := SaveLastRequest(rb); err != nil {
		t.Fatalf("SaveLastRequest failed: %v", err)
	}

	lastReq, err := LoadLastRequest()
	if err != nil {
		t.Fatalf("LoadLastRequest failed: %v", err)
	}

	if lastReq.URL != rb.URL {
		t.Errorf("URL = %q, want %q", lastReq.URL, rb.URL)
	}
	if lastReq.Method != rb.Method {
		t.Errorf("Method = %q, want %q", lastReq.Method, rb.Method)
	}
	if lastReq.Body != rb.Body {
		t.Errorf("Body = %q, want %q", lastReq.Body, rb.Body)
	}
	if lastReq.BodyFormat != rb.BodyFormat {
		t.Errorf("BodyFormat = %q, want %q", lastReq.BodyFormat, rb.BodyFormat)
	}
	if lastReq.Headers["Authorization"] != rb.Headers["Authorization"] {
		t.Errorf("Authorization header = %q, want %q", lastReq.Headers["Authorization"], rb.Headers["Authorization"])
	}
	if lastReq.QueryParams["include"] != rb.QueryParams["include"] {
		t.Errorf("include param = %q, want %q", lastReq.QueryParams["include"], rb.QueryParams["include"])
	}
	if lastReq.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
}

func TestApplyLastRequest(t *testing.T) {
	lastReq := &LastRequestJSON{
		URL:         "https://api.example.com/users",
		Method:      "PUT",
		Headers:     map[string]string{"X-Custom": "value"},
		QueryParams: map[string]string{"page": "1"},
		Body:        `{"key":"val"}`,
		BodyFormat:  "json",
	}

	rb := NewRequestBuilder()
	rb.ApplyLastRequest(lastReq)

	if rb.URL != lastReq.URL {
		t.Errorf("URL = %q, want %q", rb.URL, lastReq.URL)
	}
	if rb.Method != lastReq.Method {
		t.Errorf("Method = %q, want %q", rb.Method, lastReq.Method)
	}
	if rb.Headers["X-Custom"] != "value" {
		t.Errorf("X-Custom header = %q, want %q", rb.Headers["X-Custom"], "value")
	}
	if rb.QueryParams["page"] != "1" {
		t.Errorf("page param = %q, want %q", rb.QueryParams["page"], "1")
	}
	if rb.Body != lastReq.Body {
		t.Errorf("Body = %q, want %q", rb.Body, lastReq.Body)
	}
	if rb.BodyFormat != lastReq.BodyFormat {
		t.Errorf("BodyFormat = %q, want %q", rb.BodyFormat, lastReq.BodyFormat)
	}
}

func TestMustLoadLastRequest(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "hsp-test-must-load")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)
	os.Setenv("HOME", testDir)

	_, err := MustLoadLastRequest()
	if err == nil {
		t.Fatal("Expected error when no last request exists")
	}
	if err.Error() != "no previous request found" {
		t.Errorf("Expected 'no previous request found', got %q", err.Error())
	}
}

// ==============================
// Variables tests (variables.go)
// ==============================

func TestResolveVariablesMissing(t *testing.T) {
	vars := map[string]string{"EXISTS": "hello"}

	result, missing := ResolveVariables("{{EXISTS}} and {{MISSING}}", vars)
	if result != "hello and {{MISSING}}" {
		t.Errorf("result = %q, want %q", result, "hello and {{MISSING}}")
	}
	if len(missing) != 1 || missing[0] != "MISSING" {
		t.Errorf("missing = %v, want [MISSING]", missing)
	}
}

func TestResolveAllWithMissing(t *testing.T) {
	req := &RequestBuilder{
		URL:         "{{BASE_URL}}/users/{{ID}}",
		Headers:     map[string]string{"Authorization": "Bearer {{TOKEN}}"},
		QueryParams: map[string]string{"format": "{{FORMAT}}"},
		Body:        `{"url": "{{BASE_URL}}", "id": "{{ID}}"}`,
	}

	vars := map[string]string{"BASE_URL": "https://api.example.com"}

	missing := ResolveAll(req, vars)

	foundMissing := make(map[string]bool)
	for _, m := range missing {
		foundMissing[m] = true
	}

	if !foundMissing["ID"] {
		t.Error("ID should be in missing vars")
	}
	if !foundMissing["TOKEN"] {
		t.Error("TOKEN should be in missing vars")
	}
	if !foundMissing["FORMAT"] {
		t.Error("FORMAT should be in missing vars")
	}
	if foundMissing["BASE_URL"] {
		t.Error("BASE_URL should not be in missing vars")
	}
}

func TestResolveAllWithBodyOnly(t *testing.T) {
	req := &RequestBuilder{
		URL:         "https://api.example.com/users",
		Body:        `{"name": "{{NAME}}", "email": "{{EMAIL}}"}`,
		Headers:     map[string]string{"Content-Type": "application/json"},
		QueryParams: map[string]string{},
	}

	vars := map[string]string{"NAME": "John"}

	missing := ResolveAll(req, vars)

	if req.Body != `{"name": "John", "email": "{{EMAIL}}"}` {
		t.Errorf("Body = %q, want %q", req.Body, `{"name": "John", "email": "{{EMAIL}}"}`)
	}

	foundMissing := make(map[string]bool)
	for _, m := range missing {
		foundMissing[m] = true
	}

	if !foundMissing["EMAIL"] {
		t.Error("EMAIL should be in missing vars")
	}
	if foundMissing["NAME"] {
		t.Error("NAME should not be in missing vars")
	}
}

// ==============================
// Priority tests (priority.go)
// ==============================

func TestGetPriority(t *testing.T) {
	tests := []struct {
		key  string
		want int
	}{
		{"id", PriorityHigh},
		{"name", PriorityHigh},
		{"email", PriorityHigh},
		{"token", PriorityHigh},
		{"secret", PriorityHigh},
		{"password", PriorityHigh},
		{"status", PriorityHigh},
		{"error", PriorityHigh},
		{"data", PriorityHigh},
		{"result", PriorityHigh},
		{"access_token", PriorityHigh},
		{"created_at", PriorityLow},
		{"updated_at", PriorityLow},
		{"timestamp", PriorityLow},
		{"__v", PriorityLow},
		{"limit", PriorityLow},
		{"offset", PriorityLow},
		{"page", PriorityLow},
		{"total", PriorityLow},
		{"cursor", PriorityLow},
		{"description", PriorityMedium},
		{"title", PriorityMedium},
		{"count", PriorityMedium},
	}

	for _, tt := range tests {
		got := GetPriority(tt.key)
		if got != tt.want {
			t.Errorf("GetPriority(%q) = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestIsImportant(t *testing.T) {
	if !IsImportant("id") {
		t.Error("IsImportant('id') should be true")
	}
	if !IsImportant("token") {
		t.Error("IsImportant('token') should be true")
	}
	if !IsImportant("status") {
		t.Error("IsImportant('status') should be true")
	}
	if !IsImportant("message") {
		t.Error("IsImportant('message') should be true")
	}
	if IsImportant("description") {
		t.Error("IsImportant('description') should be false")
	}
	if IsImportant("count") {
		t.Error("IsImportant('count') should be false")
	}
}

func TestIsLowPriority(t *testing.T) {
	if !IsLowPriority("_id") {
		t.Error("IsLowPriority('_id') should be true")
	}
	if !IsLowPriority("created_at") {
		t.Error("IsLowPriority('created_at') should be true")
	}
	if !IsLowPriority("page") {
		t.Error("IsLowPriority('page') should be true")
	}
	if !IsLowPriority("limit") {
		t.Error("IsLowPriority('limit') should be true")
	}
	if IsLowPriority("id") {
		t.Error("IsLowPriority('id') should be false")
	}
	if IsLowPriority("name") {
		t.Error("IsLowPriority('name') should be false")
	}
}

func TestDetectPriority(t *testing.T) {
	if DetectPriority("id") != PriorityHigh {
		t.Error("DetectPriority('id') should be PriorityHigh")
	}
	if DetectPriority("created_at") != PriorityLow {
		t.Error("DetectPriority('created_at') should be PriorityLow")
	}
	if DetectPriority("description") != PriorityMedium {
		t.Error("DetectPriority('description') should be PriorityMedium")
	}
}

// ==============================
// Profile tests (profiles.go)
// ==============================

func TestProfilePath(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "hsp-test-profile-path")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)
	os.Setenv("HOME", testDir)

	expected := filepath.Join(testDir, ".hsp", "profiles", "mytest.json")
	got := profilePath("mytest")

	if got != expected {
		t.Errorf("profilePath('mytest') = %q, want %q", got, expected)
	}
}

func TestSaveAndLoadProfile(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "hsp-test-profile-save")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)
	os.Setenv("HOME", testDir)

	profile := &Profile{
		Name:        "testprofile",
		URL:         "https://api.example.com/users",
		Method:      "GET",
		Headers:     map[string]string{"Accept": "application/json"},
		QueryParams: map[string]string{"page": "1"},
		Body:        "",
		BodyFormat:  "",
		CreatedAt:   "2026-01-01T00:00:00Z",
	}

	if err := SaveProfile(profile); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	loaded, err := LoadProfile("testprofile")
	if err != nil {
		t.Fatalf("LoadProfile failed: %v", err)
	}

	if loaded.Name != "testprofile" {
		t.Errorf("Name = %q, want %q", loaded.Name, "testprofile")
	}
	if loaded.URL != "https://api.example.com/users" {
		t.Errorf("URL = %q, want %q", loaded.URL, "https://api.example.com/users")
	}
	if loaded.Method != "GET" {
		t.Errorf("Method = %q, want %q", loaded.Method, "GET")
	}
	if loaded.Headers["Accept"] != "application/json" {
		t.Errorf("Accept header = %q, want %q", loaded.Headers["Accept"], "application/json")
	}
	if loaded.QueryParams["page"] != "1" {
		t.Errorf("page param = %q, want %q", loaded.QueryParams["page"], "1")
	}
	if loaded.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
	if loaded.UpdatedAt == "" {
		t.Error("UpdatedAt should not be empty")
	}
}

func TestDeleteProfile(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "hsp-test-profile-delete")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)
	os.Setenv("HOME", testDir)

	profile := &Profile{
		Name:   "deleteprofile",
		URL:    "https://api.example.com",
		Method: "GET",
	}

	if err := SaveProfile(profile); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	if err := DeleteProfile("deleteprofile"); err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	_, err := LoadProfile("deleteprofile")
	if err == nil {
		t.Error("Expected error when loading deleted profile")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected not-exist error, got %v", err)
	}
}

// ==============================
// Request tests (request.go)
// ==============================

func TestRenderRequestPreview(t *testing.T) {
	req := &RequestBuilder{
		URL:    "https://api.example.com/users",
		Method: "POST",
	}

	result := RenderRequestPreview(req)

	if !strings.Contains(result, "POST") {
		t.Error("RenderRequestPreview should contain method POST")
	}
	if !strings.Contains(result, "https://api.example.com/users") {
		t.Error("RenderRequestPreview should contain URL")
	}
	if !strings.HasPrefix(result, "+") {
		t.Error("RenderRequestPreview should start with border")
	}
	if !strings.HasSuffix(result, "+") {
		t.Error("RenderRequestPreview should end with border")
	}
}

func TestGetStatusMessage(t *testing.T) {
	rb := &RequestBuilder{}

	tests := []struct {
		code int
		want string
	}{
		{200, "OK"},
		{201, "Created"},
		{204, "No Content"},
		{301, "Moved Permanently"},
		{302, "Found"},
		{304, "Not Modified"},
		{400, "Bad Request"},
		{401, "Unauthorized"},
		{403, "Forbidden"},
		{404, "Not Found"},
		{500, "Internal Server Error"},
		{502, "Bad Gateway"},
		{503, "Service Unavailable"},
		{299, "OK"},
		{399, "Redirect"},
		{499, "Client Error"},
		{599, "Server Error"},
		{999, "Unknown"},
	}

	for _, tt := range tests {
		got := rb.GetStatusMessage(tt.code)
		if got != tt.want {
			t.Errorf("GetStatusMessage(%d) = %q, want %q", tt.code, got, tt.want)
		}
	}
}

// ==============================
// Config coverage (config_test.go append)
// ==============================

func TestGetActiveEnvCustom(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "hsp-test-custom-env")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)
	os.Setenv("HOME", testDir)
	config = nil

	_, err := GetActiveEnv()
	if err != nil {
		t.Fatalf("GetActiveEnv failed with default: %v", err)
	}
}
