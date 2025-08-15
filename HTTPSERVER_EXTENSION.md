# DuckDB HTTP Server Extension

This document describes how to configure and use the DuckDB HTTP Server extension with duckdb_featureserv.

## Overview

The DuckDB HTTP Server extension allows you to enable a direct HTTP interface to your DuckDB instance, running alongside the duckdb_featureserv application. This extension is provided by the [duckdb-extension-httpserver](https://github.com/Query-farm/duckdb-extension-httpserver) project.

## Configuration

The HTTP server extension is controlled by the following configuration options in the `[DuckDB]` section:

### EnableHttpServer

- **Type**: Boolean
- **Default**: `false`
- **Description**: Controls whether the DuckDB HTTP server extension is enabled. Must be explicitly set to `true` to activate.

### Port

- **Type**: Integer
- **Default**: `9001`
- **Description**: The port number on which the DuckDB HTTP server will listen.

### ApiKey

- **Type**: String
- **Default**: Empty string
- **Description**: Optional API key for authentication. If empty, the HTTP server runs without authentication.

## Configuration Examples

### Basic Configuration (No Authentication)

```toml
[DuckDB]
EnableHttpServer = true
Port = 9001
```

### Configuration with API Key Authentication

```toml
[DuckDB]
EnableHttpServer = true
Port = 8080
ApiKey = "your_secure_api_key_here"
```

### Environment Variables

You can also configure these settings using environment variables with the `DUCKDBFS_` prefix:

```bash
export DUCKDBFS_DUCKDB_ENABLEHTTPSERVER=true
export DUCKDBFS_DUCKDB_PORT=8080
export DUCKDBFS_DUCKDB_APIKEY=your_api_key
```

## Usage

When enabled, the HTTP server extension will:

1. **Install** the httpserver extension from the community repository
2. **Load** the extension into the DuckDB instance
3. **Start** the HTTP server on the specified port with optional authentication

### Startup Logs

You'll see log messages indicating the status:

```
INFO[0000] DuckDB HTTP server started on localhost:9001 without authentication
```

Or with authentication:

```
INFO[0000] DuckDB HTTP server started on localhost:8080 with API key authentication
```

### Error Handling

If the extension fails to install, load, or start, error messages will be logged:

```
ERROR[0000] Failed to install httpserver extension: <error details>
ERROR[0000] Failed to load httpserver extension: <error details>
ERROR[0000] Failed to start DuckDB HTTP server: <error details>
```

## Security Considerations

### API Key Authentication

- Always use a strong API key in production environments
- Keep your API key secure and don't commit it to version control
- Consider using environment variables for API key configuration in production

### Network Security

- The HTTP server binds to `localhost` by default for security
- Consider firewall rules and network access controls
- Monitor the HTTP server logs for unauthorized access attempts

### Port Selection

- Choose a port that doesn't conflict with other services
- The default port `9001` is different from the main duckdb_featureserv port (`9000`)
- Ensure your chosen port is not already in use

## HTTP Server Extension Documentation

For detailed information about the HTTP server extension API and functionality, refer to the [duckdb-extension-httpserver documentation](https://github.com/Query-farm/duckdb-extension-httpserver).

## Testing

The httpserver extension functionality includes comprehensive tests covering:

- Configuration loading from files and environment variables
- HTTP server disabled/enabled scenarios
- Authentication with and without API keys
- Error handling for installation/loading failures
- Port configuration variations

Run the tests with:

```bash
go test ./internal/conf/ -run="TestDuckDB" -v
go test ./internal/data/ -run="TestHttp" -v
```

## Troubleshooting

### Extension Not Available

If you see errors about the httpserver extension not being found:

1. Ensure you have an internet connection (required to download from community repository)
2. Check that your DuckDB version supports community extensions
3. Verify there are no firewall restrictions blocking the download

### Port Conflicts

If you see errors about port binding:

1. Check that the specified port is not already in use
2. Try a different port number
3. Ensure you have permission to bind to the chosen port

### Authentication Issues

If you have problems with API key authentication:

1. Verify the API key is properly configured
2. Check for special characters that might need escaping
3. Ensure the API key is not empty if authentication is intended
