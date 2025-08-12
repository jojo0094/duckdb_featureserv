# duckdb_featureserv

A lightweight RESTful geospatial feature server for [DuckDB](https://duckdb.org/) with [duckdb-spatial](https://github.com/duckdb/duckdb-spatial) support, written in [Go](https://golang.org/). It supports the [*OGC API - Features*](https://ogcapi.ogc.org/features/) REST API standard.

This is a refactored version of [`pg_featureserv`](https://github.com/CrunchyData/pg_featureserv) adapted to work with DuckDB's spatial extension instead of Postgres/PostGIS.

## Features

* Implements the [*OGC API - Features*](https://ogcapi.ogc.org/features/) standard.
  * Standard query parameters: `limit`, `bbox`, `bbox-crs`, property filtering, `sortby`, `crs`
  * Query parameters `filter` and `filter-crs` allow [CQL filtering](https://portal.ogc.org/files/96288), with spatial support
  * Extended query parameters: `offset`, `properties`, `transform`, `precision`, `groupby`
* Data responses are formatted in JSON and [GeoJSON](https://www.rfc-editor.org/rfc/rfc7946.txt)
* Provides a simple HTML user interface, with web maps to view spatial data
* Uses the power of DuckDB to provide fast analytical queries
  and efficient spatial data processing.
  * Feature collections are defined by database tables with spatial columns
  * Filters are executed in the database using DuckDB's query engine
* Uses DuckDB Spatial extension to provide geospatial functionality:
  * Spatial filtering with geometry operations
  * Fast spatial indexing and queries
  * Marshalling feature data into GeoJSON
* Full-featured HTTP support
  * CORS support with configurable Allowed Origins
  * GZIP response encoding
  * HTTP and HTTPS support

For a full list of software capabilities see [FEATURES](FEATURES.md).

## Documentation

* [FEATURES](FEATURES.md) - full list of software capabilities
* [API](API.md) - summary of the web service API

### Relevant Standards

* [*OGC API - Features - Part 1: Core*](http://docs.ogc.org/is/17-069r3/17-069r3.html)
* [*OGC API - Features - Part 2: Coordinate Reference Systems by Reference*](https://docs.ogc.org/is/18-058/18-058.html)
* [**DRAFT** *OGC API - Features - Part 3: Filtering*](http://docs.ogc.org/DRAFTS/19-079r1.html)
* [**DRAFT** *Common Query Language (CQL2)*](https://docs.ogc.org/DRAFTS/21-065.html)
* [*GeoJSON*](https://www.rfc-editor.org/rfc/rfc7946.txt)

## Build from Source

`duckdb_featureserv` requires Go 1.24+ to support the latest DuckDB driver.

In the following, replace version `<VERSION>` with the `duckdb_featureserv` version you are building against.

### Without a Go environment

Without `go` installed, you can build `duckdb_featureserv` in a docker image:

* Download or clone this repository 
* Run the following command in `duckdb_featureserv/`:
  ```bash
  make APPVERSION=<VERSION> clean build-in-docker
  ```

### In Go environment

* Download or clone this repository
* To build the executable, run the following commands:
  ```bash
  cd duckdb_featureserv/
  go build
  ```

* This creates a `duckdb_featureserv` executable in the application directory
* (Optional) Run the unit tests using `go test ./...`

### Docker image of `duckdb_featureserv`

#### Build the image

```bash
make APPVERSION=<VERSION> clean docker
```

#### Run the image

To run using an image built above, and mount a local, pre-made DuckDB database in the container:

```bash
docker run --rm -dt -v "$PWD/database.duckdb:/data/database.duckdb" -e DUCKDB_PATH=/data/database.duckd -p 9000:9000 tobilg/duckdb-featureserv:<VERSION>
```

## Configure the service

The [configuration file](config/duckdb_featureserv.toml.example) is automatically read from the following locations, if it exists:

* In the system configuration directory, at `/etc/duckdb_featureserv.toml`
* Relative to the directory from which the program is run, `./config/duckdb_featureserv.toml`
* In a root volume at `/config/duckdb_featureserv.toml`

To specify a configuration file directly use the `--config` commandline parameter.
In this case configuration files in other locations are ignored.

### Configuration Using Environment Variables

To set the database connection the environment variable `DUCKDB_PATH`
can be used to specify the path to a DuckDB database file:
```bash
export DUCKDB_PATH="/path/to/database.db"
```

To specify which table to serve, use the `DUCKDB_TABLE` environment variable:
```bash
export DUCKDB_TABLE="my_spatial_table"
```

Other parameters in the configuration file can be over-ridden in the environment.
Prepend the upper-cased parameter name with `DUCKDBFS_section_` to set the value.
For example, to change the HTTP port and service title:
```bash
export DUCKDBFS_SERVER_HTTPPORT=8889
export DUCKDBFS_METADATA_TITLE="My DuckDB FeatureServ"
```

### SSL
For SSL support, a server **private key** and an **authority certificate** are needed.
For testing purposes you can generate a **self-signed key/cert pair** using `openssl`:
```bash
openssl req  -nodes -new -x509  -keyout server.key -out server.crt
```
These are set in the configuration file:
```
TlsServerCertificateFile = "/path/server.crt"
TlsServerPrivateKeyFile = "/path/server.key"
```

## Run the service

* Change to the application directory:
  * `cd duckdb_featureserv/`
* Start the server with a DuckDB database:
  * `./duckdb_featureserv --duckdb /path/to/your/database.db --table your_spatial_table`
* Or set environment variables:
  * `export DUCKDB_PATH="/path/to/your/database.db"`
  * `export DUCKDB_TABLE="your_spatial_table"`
  * `./duckdb_featureserv`
* Open the service home page in a browser:
  * `http://localhost:9000/home.html`

### Command-line options

* `-?` - show command usage
* `--config file.toml` - specify configuration file to use
* `--debug` - set logging level to TRACE (can also be set in config file)
* `--devel` - run in development mode (e.g. HTML templates reloaded every query)
* `--test` - run in test mode, with an internal catalog of tables and data
* `--version` - display the version number
* `--duckdb path` - specify path to DuckDB database file
* `--table name` - specify name of spatial table to serve (optional; if not specified, all tables with geometry columns will be served)

## Testing

### Automated Testing

Run the comprehensive test suite:

```bash
# Create test database and run all tests
./testing/test_duckdb_spatial.sh

# Or run API endpoint tests (requires server to be running)
./duckdb_featureserv --duckdb test_spatial.duckdb &
./testing/test_api_endpoints.sh
```

### Manual Testing

```bash
# Create test database with spatial data
duckdb test_spatial.duckdb < testing/duckdb_test.sql

# Start server
./duckdb_featureserv --duckdb test_spatial.duckdb

# Test collections endpoint
curl "http://localhost:9000/collections"

# Test features endpoint
curl "http://localhost:9000/collections/test_geom/items"
```

See `testing/duckdb_test.md` for comprehensive test case documentation.

## Troubleshooting

To get detailed information about service operation
run with the `--debug` commandline parameter.
```sh
./duckdb_featureserv --debug
```
Debugging can also be enabled via the configuration file (`Server.Debug=true`).

## Requests Overview

Features are identified by a _collection name_ and _feature id_ pair.

The default response is in JSON/GeoJSON format.
Append `.html` to the request path to see the UI page for the resource.
In a web browser, to request a JSON response append `.json` to the path (which overrides the browser `Accept` header).

The example requests assume that the service is running locally and configured
to listen on port 9000.

- Landing page (HTML or JSON): http://localhost:9000/
- Landing page (HTML): http://localhost:9000/index.html
- Landing page (JSON): http://localhost:9000/index.json
- OpenAPI definition: http://localhost:9000/api
- OpenAPI test UI: http://localhost:9000/api.html
- Conformance: http://localhost:9000/conformance
- Collections: http://localhost:9000/collections
- Collections UI: http://localhost:9000/collections.html
- Feature collection metadata: http://localhost:9000/collections/{name}
- Feature collection UI: http://localhost:9000/collections/{name}.html
- Features from a single feature collection: http://localhost:9000/collections/{name}/items
- Features from a single feature collection (Map UI): http://localhost:9000/collections/{name}/items.html
- Single feature from a feature collection: http://localhost:9000/collections/{name}/items/{featureid}
- Functions (JSON): http://localhost:9000/functions
- Functions UI: http://localhost:9000/functions.html
- Function metadata: http://localhost:9000/functions/{name}
- Function UI: http://localhost:9000/functions/{name}.html
- Features from a function (JSON): http://localhost:9000/functions/{name}/items
- Features from a function (Map UI): http://localhost:9000/functions/{name}/items.html

See [API Summary](API.md) for a summary of the web service API.
