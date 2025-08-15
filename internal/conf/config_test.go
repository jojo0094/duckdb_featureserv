package conf

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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/spf13/viper"
)

// TestTableIncludesEnvironmentVariable tests that TableIncludes can be set via environment variable
func TestTableIncludesEnvironmentVariable(t *testing.T) {
	// Clear any existing environment variables
	defer clearConfigEnvVars()

	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "Single table",
			envValue: "public.table1",
			expected: []string{"public.table1"},
		},
		{
			name:     "Multiple tables",
			envValue: "public,schema1.table1,table2",
			expected: []string{"public", "schema1.table1", "table2"},
		},
		{
			name:     "Empty value",
			envValue: "",
			expected: []string{},
		},
		{
			name:     "Schema only",
			envValue: "public",
			expected: []string{"public"},
		},
		{
			name:     "Complex table names",
			envValue: "my_schema.my_table,another_schema.complex_table_name",
			expected: []string{"my_schema.my_table", "another_schema.complex_table_name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			clearConfigEnvVars()

			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("DUCKDBFS_DATABASE_TABLEINCLUDES", tt.envValue)
			}

			// Reset viper for clean state
			viper.Reset()

			// Initialize config
			InitConfig("", false)

			// Check result
			equals(t, tt.expected, Configuration.Database.TableIncludes, "TableIncludes")

			// Clean up
			clearConfigEnvVars()
		})
	}
}

// TestTableExcludesEnvironmentVariable tests that TableExcludes can be set via environment variable
func TestTableExcludesEnvironmentVariable(t *testing.T) {
	// Clear any existing environment variables
	defer clearConfigEnvVars()

	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "Single table exclusion",
			envValue: "private.secrets",
			expected: []string{"private.secrets"},
		},
		{
			name:     "Multiple table exclusions",
			envValue: "private,temp,logs.debug",
			expected: []string{"private", "temp", "logs.debug"},
		},
		{
			name:     "Empty value",
			envValue: "",
			expected: []string{},
		},
		{
			name:     "Schema exclusion",
			envValue: "temp",
			expected: []string{"temp"},
		},
		{
			name:     "Complex exclusion patterns",
			envValue: "temp_schema,staging.test_data,logs.debug_logs",
			expected: []string{"temp_schema", "staging.test_data", "logs.debug_logs"},
		},
		{
			name:     "System and private exclusions",
			envValue: "system,private,internal.migrations",
			expected: []string{"system", "private", "internal.migrations"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			clearConfigEnvVars()

			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("DUCKDBFS_DATABASE_TABLEEXCLUDES", tt.envValue)
			}

			// Reset viper for clean state
			viper.Reset()

			// Initialize config
			InitConfig("", false)

			// Check result
			equals(t, tt.expected, Configuration.Database.TableExcludes, "TableExcludes")

			// Clean up
			clearConfigEnvVars()
		})
	}
}

// TestConfigFileOverriddenByEnvironment tests that environment variables take precedence over config file
func TestConfigFileOverriddenByEnvironment(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	// Create a temporary config file
	configContent := `
[Database]
TableIncludes = ["file_table1", "file_table2"]
TableExcludes = ["file_exclude"]
`

	tempDir, err := os.MkdirTemp("", "duckdb_featureserv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "test_config.toml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Set environment variables that should override config file
	os.Setenv("DUCKDBFS_DATABASE_TABLEINCLUDES", "env_table1,env_table2")
	os.Setenv("DUCKDBFS_DATABASE_TABLEEXCLUDES", "env_exclude")
	defer func() {
		os.Unsetenv("DUCKDBFS_DATABASE_TABLEINCLUDES")
		os.Unsetenv("DUCKDBFS_DATABASE_TABLEEXCLUDES")
	}()

	viper.Reset()
	InitConfig(configFile, false)

	// Environment variables should take precedence
	expectedIncludes := []string{"env_table1", "env_table2"}
	expectedExcludes := []string{"env_exclude"}

	equals(t, expectedIncludes, Configuration.Database.TableIncludes, "TableIncludes from env")
	equals(t, expectedExcludes, Configuration.Database.TableExcludes, "TableExcludes from env")
}

// TestConfigFileOnly tests that config file values are used when no environment variables are set
func TestConfigFileOnly(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	configContent := `
[Database]
TableIncludes = ["config_table1", "config_table2"]
TableExcludes = ["config_exclude"]
`

	tempDir, err := os.MkdirTemp("", "duckdb_featureserv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "test_config.toml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	viper.Reset()
	InitConfig(configFile, false)

	expectedIncludes := []string{"config_table1", "config_table2"}
	expectedExcludes := []string{"config_exclude"}

	equals(t, expectedIncludes, Configuration.Database.TableIncludes, "TableIncludes from config")
	equals(t, expectedExcludes, Configuration.Database.TableExcludes, "TableExcludes from config")
}

// TestDefaultValues tests that default values are used when no config file or environment variables are set
func TestDefaultValues(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	viper.Reset()
	InitConfig("", false)

	// Should have empty slices as defaults
	equals(t, []string{}, Configuration.Database.TableIncludes, "Default TableIncludes")
	equals(t, []string{}, Configuration.Database.TableExcludes, "Default TableExcludes")
}

// TestEnvironmentVariableFormat tests various formats for the environment variable
func TestEnvironmentVariableFormat(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "No spaces",
			envValue: "table1,table2,table3",
			expected: []string{"table1", "table2", "table3"},
		},
		{
			name:     "With spaces (Viper doesn't trim)",
			envValue: "table1, table2 , table3",
			expected: []string{"table1", " table2 ", " table3"},
		},
		{
			name:     "Single item",
			envValue: "single_table",
			expected: []string{"single_table"},
		},
		{
			name:     "Mixed schema.table and table only",
			envValue: "schema1.table1,table2,schema2.table3",
			expected: []string{"schema1.table1", "table2", "schema2.table3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnvVars()
			os.Setenv("DUCKDBFS_DATABASE_TABLEINCLUDES", tt.envValue)

			viper.Reset()
			InitConfig("", false)

			// Check that configuration matches expected values
			equals(t, tt.expected, Configuration.Database.TableIncludes, "TableIncludes")
		})
	}
}

// TestDuckDBConfigDefaults tests that DuckDB configuration has correct default values
func TestDuckDBConfigDefaults(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	viper.Reset()
	InitConfig("", false)

	// Test default values for DuckDB configuration
	equals(t, false, Configuration.DuckDB.EnableHttpServer, "Default EnableHttpServer")
	equals(t, 9001, Configuration.DuckDB.Port, "Default Port")
	equals(t, "", Configuration.DuckDB.ApiKey, "Default ApiKey")
}

// TestDuckDBConfigFromFile tests that DuckDB configuration can be loaded from config file
func TestDuckDBConfigFromFile(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	configContent := `
[DuckDB]
EnableHttpServer = true
Port = 8080
ApiKey = "test_key_123"
`

	tempDir, err := os.MkdirTemp("", "duckdb_featureserv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "test_config.toml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	viper.Reset()
	InitConfig(configFile, false)

	equals(t, true, Configuration.DuckDB.EnableHttpServer, "EnableHttpServer from config")
	equals(t, 8080, Configuration.DuckDB.Port, "Port from config")
	equals(t, "test_key_123", Configuration.DuckDB.ApiKey, "ApiKey from config")
}

// TestDuckDBConfigFromEnvironment tests that DuckDB configuration can be loaded from environment variables
func TestDuckDBConfigFromEnvironment(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	// Set environment variables
	os.Setenv("DUCKDBFS_DUCKDB_ENABLEHTTPSERVER", "true")
	os.Setenv("DUCKDBFS_DUCKDB_PORT", "7777")
	os.Setenv("DUCKDBFS_DUCKDB_APIKEY", "env_api_key")
	defer func() {
		os.Unsetenv("DUCKDBFS_DUCKDB_ENABLEHTTPSERVER")
		os.Unsetenv("DUCKDBFS_DUCKDB_PORT")
		os.Unsetenv("DUCKDBFS_DUCKDB_APIKEY")
	}()

	viper.Reset()
	InitConfig("", false)

	equals(t, true, Configuration.DuckDB.EnableHttpServer, "EnableHttpServer from env")
	equals(t, 7777, Configuration.DuckDB.Port, "Port from env")
	equals(t, "env_api_key", Configuration.DuckDB.ApiKey, "ApiKey from env")
}

// TestDuckDBConfigEnvironmentOverridesFile tests that environment variables override config file
func TestDuckDBConfigEnvironmentOverridesFile(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	// Create config file with one set of values
	configContent := `
[DuckDB]
EnableHttpServer = false
Port = 8080
ApiKey = "file_key"
`

	tempDir, err := os.MkdirTemp("", "duckdb_featureserv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "test_config.toml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Set environment variables with different values
	os.Setenv("DUCKDBFS_DUCKDB_ENABLEHTTPSERVER", "true")
	os.Setenv("DUCKDBFS_DUCKDB_PORT", "9999")
	os.Setenv("DUCKDBFS_DUCKDB_APIKEY", "env_override_key")
	defer func() {
		os.Unsetenv("DUCKDBFS_DUCKDB_ENABLEHTTPSERVER")
		os.Unsetenv("DUCKDBFS_DUCKDB_PORT")
		os.Unsetenv("DUCKDBFS_DUCKDB_APIKEY")
	}()

	viper.Reset()
	InitConfig(configFile, false)

	// Environment variables should take precedence
	equals(t, true, Configuration.DuckDB.EnableHttpServer, "EnableHttpServer from env override")
	equals(t, 9999, Configuration.DuckDB.Port, "Port from env override")
	equals(t, "env_override_key", Configuration.DuckDB.ApiKey, "ApiKey from env override")
}

// TestDuckDBConfigPartialOverride tests that only set environment variables override config file
func TestDuckDBConfigPartialOverride(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	configContent := `
[DuckDB]
EnableHttpServer = true
Port = 8080
ApiKey = "file_key"
`

	tempDir, err := os.MkdirTemp("", "duckdb_featureserv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "test_config.toml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Only override the port via environment variable
	os.Setenv("DUCKDBFS_DUCKDB_PORT", "5555")
	defer os.Unsetenv("DUCKDBFS_DUCKDB_PORT")

	viper.Reset()
	InitConfig(configFile, false)

	// Port should be from env, others from file
	equals(t, true, Configuration.DuckDB.EnableHttpServer, "EnableHttpServer from config")
	equals(t, 5555, Configuration.DuckDB.Port, "Port from env")
	equals(t, "file_key", Configuration.DuckDB.ApiKey, "ApiKey from config")
}

// Helper function to clear all configuration-related environment variables
func clearConfigEnvVars() {
	envVars := []string{
		"DUCKDBFS_DATABASE_TABLEINCLUDES",
		"DUCKDBFS_DATABASE_TABLEEXCLUDES",
		"DUCKDBFS_DATABASE_PATH",
		"DUCKDBFS_DATABASE_TABLENAME",
		"DUCKDBFS_SERVER_HTTPPORT",
		"DUCKDBFS_SERVER_DEBUG",
		"DUCKDBFS_DUCKDB_ENABLEHTTPSERVER",
		"DUCKDBFS_DUCKDB_PORT",
		"DUCKDBFS_DUCKDB_APIKEY",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// Also clear the global Configuration variable
	Configuration = Config{}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}, msg string) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("%s:%d: %s - expected: %#v; got: %#v\n", filepath.Base(file), line, msg, exp, act)
		tb.FailNow()
	}
}
