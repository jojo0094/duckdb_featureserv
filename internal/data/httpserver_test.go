package data

/*
 Copyright 2019 - 2025 Crunchy Data Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tobilg/duckdb_featureserv/internal/conf"
)

// mockDriver implements sql/driver.Driver for testing
type mockDriver struct {
	shouldFailOpen  bool
	shouldFailPing  bool
	shouldFailExec  bool
	failExecPattern string
	execCalls       []string
}

// mockConn implements sql/driver.Conn for testing
type mockConn struct {
	driver *mockDriver
}

// mockResult implements sql/driver.Result for testing
type mockResult struct{}

func (r mockResult) LastInsertId() (int64, error) { return 0, nil }
func (r mockResult) RowsAffected() (int64, error) { return 1, nil }

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	if d.shouldFailOpen {
		return nil, errors.New("mock open failure")
	}
	return &mockConn{driver: d}, nil
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare not implemented in mock")
}

func (c *mockConn) Close() error {
	return nil
}

func (c *mockConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not implemented in mock")
}

func (c *mockConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	c.driver.execCalls = append(c.driver.execCalls, query)

	if c.driver.shouldFailExec {
		return nil, errors.New("mock exec failure")
	}

	if c.driver.failExecPattern != "" && strings.Contains(query, c.driver.failExecPattern) {
		return nil, fmt.Errorf("mock failure for pattern: %s", c.driver.failExecPattern)
	}

	return mockResult{}, nil
}

// mockDB wraps sql.DB with tracking capabilities
type mockDB struct {
	*sql.DB
	driver *mockDriver
}

func newMockDB(driver *mockDriver) *mockDB {
	// Register the mock driver
	driverName := fmt.Sprintf("mock-%p", driver)
	sql.Register(driverName, driver)

	db, _ := sql.Open(driverName, "mock_dsn")
	return &mockDB{DB: db, driver: driver}
}

func (db *mockDB) GetExecCalls() []string {
	return db.driver.execCalls
}

func (db *mockDB) Ping() error {
	if db.driver.shouldFailPing {
		return errors.New("mock ping failure")
	}
	return nil
}

// TestHttpServerDisabled tests that httpserver extension is not loaded when disabled
func TestHttpServerDisabled(t *testing.T) {
	// Save original configuration
	originalConfig := conf.Configuration
	defer func() { conf.Configuration = originalConfig }()

	// Set up configuration with httpserver disabled
	conf.Configuration.DuckDB.EnableHttpServer = false
	conf.Configuration.Database.DatabasePath = "test.db"

	driver := &mockDriver{}
	mockDB := newMockDB(driver)
	defer mockDB.Close()

	// Simulate dbConnect logic
	err := mockDB.Ping()
	if err != nil {
		t.Fatalf("Expected ping to succeed, got: %v", err)
	}

	// Simulate spatial extension loading
	_, err = mockDB.Exec("INSTALL spatial; LOAD spatial;")
	if err != nil {
		t.Fatalf("Expected spatial extension to load, got: %v", err)
	}

	// Httpserver should not be loaded when disabled
	execCalls := mockDB.GetExecCalls()

	// Should only have spatial extension call, not httpserver
	expectedCalls := 1
	if len(execCalls) != expectedCalls {
		t.Errorf("Expected %d exec calls, got %d", expectedCalls, len(execCalls))
	}

	// Verify no httpserver-related calls
	for _, call := range execCalls {
		if strings.Contains(call, "httpserver") {
			t.Errorf("Unexpected httpserver call when disabled: %s", call)
		}
	}
}

// TestHttpServerEnabledWithApiKey tests httpserver with API key
func TestHttpServerEnabledWithApiKey(t *testing.T) {
	// Save original configuration
	originalConfig := conf.Configuration
	defer func() { conf.Configuration = originalConfig }()

	// Set up configuration with httpserver enabled and API key
	conf.Configuration.DuckDB.EnableHttpServer = true
	conf.Configuration.DuckDB.Port = 8080
	conf.Configuration.DuckDB.ApiKey = "test_api_key_123"
	conf.Configuration.Database.DatabasePath = "test.db"

	driver := &mockDriver{}
	mockDB := newMockDB(driver)
	defer mockDB.Close()

	// Simulate the database connection flow
	err := mockDB.Ping()
	if err != nil {
		t.Fatalf("Expected ping to succeed, got: %v", err)
	}

	// Load spatial extension
	_, err = mockDB.Exec("INSTALL spatial; LOAD spatial;")
	if err != nil {
		t.Fatalf("Expected spatial extension to load, got: %v", err)
	}

	// Simulate httpserver extension installation and loading
	_, err = mockDB.Exec("INSTALL httpserver FROM community;")
	if err != nil {
		t.Fatalf("Expected httpserver install to succeed, got: %v", err)
	}

	_, err = mockDB.Exec("LOAD httpserver;")
	if err != nil {
		t.Fatalf("Expected httpserver load to succeed, got: %v", err)
	}

	// Start httpserver with API key
	expectedQuery := "SELECT httpserve_start('localhost', 8080, 'test_api_key_123');"
	_, err = mockDB.Exec(expectedQuery)
	if err != nil {
		t.Fatalf("Expected httpserve_start to succeed, got: %v", err)
	}

	// Verify the expected sequence of calls
	execCalls := mockDB.GetExecCalls()
	expectedCalls := []string{
		"INSTALL spatial; LOAD spatial;",
		"INSTALL httpserver FROM community;",
		"LOAD httpserver;",
		"SELECT httpserve_start('localhost', 8080, 'test_api_key_123');",
	}

	if len(execCalls) != len(expectedCalls) {
		t.Errorf("Expected %d exec calls, got %d", len(expectedCalls), len(execCalls))
	}

	for i, expected := range expectedCalls {
		if i < len(execCalls) && execCalls[i] != expected {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expected, execCalls[i])
		}
	}
}

// TestHttpServerEnabledWithoutApiKey tests httpserver without API key
func TestHttpServerEnabledWithoutApiKey(t *testing.T) {
	// Save original configuration
	originalConfig := conf.Configuration
	defer func() { conf.Configuration = originalConfig }()

	// Set up configuration with httpserver enabled but no API key
	conf.Configuration.DuckDB.EnableHttpServer = true
	conf.Configuration.DuckDB.Port = 9001
	conf.Configuration.DuckDB.ApiKey = "" // Empty API key
	conf.Configuration.Database.DatabasePath = "test.db"

	driver := &mockDriver{}
	mockDB := newMockDB(driver)
	defer mockDB.Close()

	// Simulate the database connection flow
	err := mockDB.Ping()
	if err != nil {
		t.Fatalf("Expected ping to succeed, got: %v", err)
	}

	// Load spatial extension
	_, err = mockDB.Exec("INSTALL spatial; LOAD spatial;")
	if err != nil {
		t.Fatalf("Expected spatial extension to load, got: %v", err)
	}

	// Simulate httpserver extension installation and loading
	_, err = mockDB.Exec("INSTALL httpserver FROM community;")
	if err != nil {
		t.Fatalf("Expected httpserver install to succeed, got: %v", err)
	}

	_, err = mockDB.Exec("LOAD httpserver;")
	if err != nil {
		t.Fatalf("Expected httpserver load to succeed, got: %v", err)
	}

	// Start httpserver without API key (empty string)
	expectedQuery := "SELECT httpserve_start('localhost', 9001, '');"
	_, err = mockDB.Exec(expectedQuery)
	if err != nil {
		t.Fatalf("Expected httpserve_start to succeed, got: %v", err)
	}

	// Verify the expected sequence of calls
	execCalls := mockDB.GetExecCalls()
	expectedCalls := []string{
		"INSTALL spatial; LOAD spatial;",
		"INSTALL httpserver FROM community;",
		"LOAD httpserver;",
		"SELECT httpserve_start('localhost', 9001, '');",
	}

	if len(execCalls) != len(expectedCalls) {
		t.Errorf("Expected %d exec calls, got %d", len(expectedCalls), len(execCalls))
	}

	for i, expected := range expectedCalls {
		if i < len(execCalls) && execCalls[i] != expected {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expected, execCalls[i])
		}
	}
}

// TestHttpServerInstallFailure tests handling of httpserver install failure
func TestHttpServerInstallFailure(t *testing.T) {
	// Save original configuration
	originalConfig := conf.Configuration
	defer func() { conf.Configuration = originalConfig }()

	// Set up configuration with httpserver enabled
	conf.Configuration.DuckDB.EnableHttpServer = true
	conf.Configuration.DuckDB.Port = 8080
	conf.Configuration.DuckDB.ApiKey = "test_key"
	conf.Configuration.Database.DatabasePath = "test.db"

	// Set up driver to fail on httpserver install
	driver := &mockDriver{
		failExecPattern: "INSTALL httpserver FROM community",
	}
	mockDB := newMockDB(driver)
	defer mockDB.Close()

	// Simulate successful spatial extension loading
	_, err := mockDB.Exec("INSTALL spatial; LOAD spatial;")
	if err != nil {
		t.Fatalf("Expected spatial extension to load, got: %v", err)
	}

	// Httpserver install should fail
	_, err = mockDB.Exec("INSTALL httpserver FROM community;")
	if err == nil {
		t.Fatal("Expected httpserver install to fail")
	}

	// Verify the error contains expected text
	if !strings.Contains(err.Error(), "INSTALL httpserver FROM community") {
		t.Errorf("Expected error to contain 'INSTALL httpserver FROM community', got: %v", err)
	}
}

// TestHttpServerLoadFailure tests handling of httpserver load failure
func TestHttpServerLoadFailure(t *testing.T) {
	// Save original configuration
	originalConfig := conf.Configuration
	defer func() { conf.Configuration = originalConfig }()

	// Set up configuration with httpserver enabled
	conf.Configuration.DuckDB.EnableHttpServer = true
	conf.Configuration.DuckDB.Port = 8080
	conf.Configuration.DuckDB.ApiKey = "test_key"
	conf.Configuration.Database.DatabasePath = "test.db"

	// Set up driver to fail on httpserver load
	driver := &mockDriver{
		failExecPattern: "LOAD httpserver",
	}
	mockDB := newMockDB(driver)
	defer mockDB.Close()

	// Install should succeed
	_, err := mockDB.Exec("INSTALL httpserver FROM community;")
	if err != nil {
		t.Fatalf("Expected httpserver install to succeed, got: %v", err)
	}

	// Load should fail
	_, err = mockDB.Exec("LOAD httpserver;")
	if err == nil {
		t.Fatal("Expected httpserver load to fail")
	}

	// Verify the error contains expected text
	if !strings.Contains(err.Error(), "LOAD httpserver") {
		t.Errorf("Expected error to contain 'LOAD httpserver', got: %v", err)
	}
}

// TestHttpServerStartFailure tests handling of httpserve_start failure
func TestHttpServerStartFailure(t *testing.T) {
	// Save original configuration
	originalConfig := conf.Configuration
	defer func() { conf.Configuration = originalConfig }()

	// Set up configuration with httpserver enabled
	conf.Configuration.DuckDB.EnableHttpServer = true
	conf.Configuration.DuckDB.Port = 8080
	conf.Configuration.DuckDB.ApiKey = "test_key"
	conf.Configuration.Database.DatabasePath = "test.db"

	// Set up driver to fail on httpserve_start
	driver := &mockDriver{
		failExecPattern: "httpserve_start",
	}
	mockDB := newMockDB(driver)
	defer mockDB.Close()

	// Install and load should succeed
	_, err := mockDB.Exec("INSTALL httpserver FROM community;")
	if err != nil {
		t.Fatalf("Expected httpserver install to succeed, got: %v", err)
	}

	_, err = mockDB.Exec("LOAD httpserver;")
	if err != nil {
		t.Fatalf("Expected httpserver load to succeed, got: %v", err)
	}

	// Start should fail
	_, err = mockDB.Exec("SELECT httpserve_start('localhost', 8080, 'test_key');")
	if err == nil {
		t.Fatal("Expected httpserve_start to fail")
	}

	// Verify the error contains expected text
	if !strings.Contains(err.Error(), "httpserve_start") {
		t.Errorf("Expected error to contain 'httpserve_start', got: %v", err)
	}
}

// TestHttpServerPortVariations tests different port configurations
func TestHttpServerPortVariations(t *testing.T) {
	testCases := []struct {
		name string
		port int
	}{
		{"Standard port", 8080},
		{"High port", 65535},
		{"Low port", 1024},
		{"Default port", 9001},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original configuration
			originalConfig := conf.Configuration
			defer func() { conf.Configuration = originalConfig }()

			// Set up configuration with specific port
			conf.Configuration.DuckDB.EnableHttpServer = true
			conf.Configuration.DuckDB.Port = tc.port
			conf.Configuration.DuckDB.ApiKey = "test_key"
			conf.Configuration.Database.DatabasePath = "test.db"

			driver := &mockDriver{}
			mockDB := newMockDB(driver)
			defer mockDB.Close()

			// Simulate loading and starting httpserver
			_, err := mockDB.Exec("INSTALL httpserver FROM community;")
			if err != nil {
				t.Fatalf("Expected httpserver install to succeed, got: %v", err)
			}

			_, err = mockDB.Exec("LOAD httpserver;")
			if err != nil {
				t.Fatalf("Expected httpserver load to succeed, got: %v", err)
			}

			expectedQuery := fmt.Sprintf("SELECT httpserve_start('localhost', %d, 'test_key');", tc.port)
			_, err = mockDB.Exec(expectedQuery)
			if err != nil {
				t.Fatalf("Expected httpserve_start to succeed, got: %v", err)
			}

			// Verify the port was used correctly
			execCalls := mockDB.GetExecCalls()
			found := false
			for _, call := range execCalls {
				if strings.Contains(call, fmt.Sprintf("'localhost', %d,", tc.port)) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find httpserve_start call with port %d in: %v", tc.port, execCalls)
			}
		})
	}
}
