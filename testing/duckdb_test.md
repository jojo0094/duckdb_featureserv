# duckdb_featureserv Test Requests

HTTP requests to test `duckdb_featureserv` capabilities using DuckDB with spatial extension.

### Usage

* Initialize the database with the DDL in `duckdb_test.sql`:
  ```bash
  duckdb test_spatial.duckdb < testing/duckdb_test.sql
  ```
* Run `duckdb_featureserv`:
  ```bash
  ./duckdb_featureserv --database-path test_spatial.duckdb
  ```
* Test endpoints using curl or paste URLs into a browser

### Automated Testing

Run the complete test suite:
```bash
./testing/test_duckdb_spatial.sh
```

## Collections Discovery

### All Collections
```
http://localhost:9000/collections
```
Expected Response: 6 collections (test_arr, test_crs, test_geom, test_json, test_names, test_srid0)

### Individual Collection Metadata
```
http://localhost:9000/collections/test_crs
```

```
http://localhost:9000/collections/test_geom
```

## Basic Feature Retrieval

### All Features from Collection
```
http://localhost:9000/collections/test_geom/items
```
Expected Response: 2 features (Point and LineString)

### Limited Features
```
http://localhost:9000/collections/test_crs/items?limit=5
```
Expected Response: 5 polygon features

### Features with Properties Selection
```
http://localhost:9000/collections/test_names/items?properties=id,colCamelCase
```
Expected Response: Features with only selected properties

## Spatial Filtering with Bounding Box

### Simple Bounding Box Filter
```
http://localhost:9000/collections/test_crs/items?bbox=1000000,400000,1010000,410000
```
Expected Response: Features intersecting the bounding box

### Larger Bounding Box
```
http://localhost:9000/collections/test_crs/items?bbox=1000000,400000,1030000,430000
```
Expected Response: More features (approximately 4 features)

### Point Data Bounding Box
```
http://localhost:9000/collections/test_json/items?bbox=0.5,0.5,2.5,2.5
```
Expected Response: Both point features

## CQL Filters

### Attribute Filters
```
http://localhost:9000/collections/test_crs/items?filter=id BETWEEN 1 AND 5
```

```
http://localhost:9000/collections/test_crs/items?filter=name LIKE '1_%'
```
Expected Response: Features with names starting with "1_"

```
http://localhost:9000/collections/test_crs/items?filter=NOT id IN (1,2,3)
```
Expected Response: Features excluding IDs 1, 2, and 3

### Data Type Specific Filters
```
http://localhost:9000/collections/test_geom/items?filter=data = 'aaa'
```
Expected Response: Point feature with data='aaa'

## Spatial CQL Operators

### Point Intersection
```
http://localhost:9000/collections/test_crs/items?filter=INTERSECTS(geom, POINT(1010000 410000))
```
Expected Response: Polygons intersecting the point

### Geometry Intersection
```
http://localhost:9000/collections/test_geom/items?filter=INTERSECTS(geom, POINT(1 1))
```
Expected Response: Features intersecting point (1,1)

### LineString Intersection
```
http://localhost:9000/collections/test_crs/items?filter=INTERSECTS(geom, LINESTRING(1000000 400000, 1010000 410000))
```
Expected Response: Polygons intersecting the line

### Spatial Distance Queries
```
http://localhost:9000/collections/test_json/items?filter=DWITHIN(geom, POINT(1.5 1.5), 1)
```
Expected Response: Points within distance 1 from (1.5, 1.5)

### Contains and Within Operations
```
http://localhost:9000/collections/test_crs/items?filter=CONTAINS(geom, POINT(1005000 405000))
```
Expected Response: Polygons containing the point

## Array Data Testing

### Array Column Retrieval
```
http://localhost:9000/collections/test_arr/items?properties=id,val_int,val_txt
```
Expected Response: Feature with array values

### Array Data with All Properties
```
http://localhost:9000/collections/test_arr/items
```
Expected Response: Complete feature with all array types (boolean[], integer[], double[], varchar[])

## JSON Data Testing

### JSON Properties
```
http://localhost:9000/collections/test_json/items?properties=id,val_json
```
Expected Response: Features with JSON data

### Complete JSON Features
```
http://localhost:9000/collections/test_json/items
```
Expected Response: Point geometries with JSON attributes

## Mixed Geometry Types

### Different Geometry Types
```
http://localhost:9000/collections/test_geom/items
```
Expected Response: Both POINT and LINESTRING geometries

### Geometry Type Filtering (if supported)
```
http://localhost:9000/collections/test_geom/items?filter=ST_GeometryType(geom) = 'POINT'
```
Expected Response: Only point features

## Coordinate Reference Systems

### Default CRS (4326)
```
http://localhost:9000/collections/test_json/items
```
Expected Response: Features in default WGS84

### Custom CRS Data
```
http://localhost:9000/collections/test_crs/items?limit=1
```
Expected Response: Features with custom coordinate system data

### SRID 0 Testing
```
http://localhost:9000/collections/test_srid0/items?limit=5
```
Expected Response: Features with no specific projection

## Error Handling

### Non-existent Collection
```
http://localhost:9000/collections/nonexistent/items
```
Expected Response: 404 Collection not found

### Invalid Filter
```
http://localhost:9000/collections/test_geom/items?filter=invalid_column = 'value'
```
Expected Response: Error or empty result

### Invalid Geometry in Filter
```
http://localhost:9000/collections/test_geom/items?filter=INTERSECTS(geom, INVALID_GEOMETRY)
```
Expected Response: Error response

## Output Formats

### JSON Format (default)
```
http://localhost:9000/collections/test_geom/items.json
```

### HTML Format
```
http://localhost:9000/collections/test_geom/items.html
```

### GeoJSON Format
```
http://localhost:9000/collections/test_geom/items
```
Expected Response: Valid GeoJSON FeatureCollection

## Performance Testing

### Large Result Set
```
http://localhost:9000/collections/test_crs/items
```
Expected Response: All 100 polygon features

### Pagination
```
http://localhost:9000/collections/test_crs/items?limit=10&offset=20
```
Expected Response: 10 features starting from the 21st

### Combined Filters and Pagination
```
http://localhost:9000/collections/test_crs/items?filter=id > 50&limit=10
```
Expected Response: 10 features with ID > 50

## Column Name Handling

### Quoted Column Names
```
http://localhost:9000/collections/test_names/items?properties=id,"colCamelCase"
```
Expected Response: Features with camelCase column

### Case Sensitivity
```
http://localhost:9000/collections/test_names/items?filter="colCamelCase" = 1
```
Expected Response: Feature with colCamelCase = 1

## API Metadata

### Service Landing Page
```
http://localhost:9000/
```

### OpenAPI Specification
```
http://localhost:9000/api
```

### Collections Metadata
```
http://localhost:9000/collections
```
Expected Response: All 6 collections with metadata

### Individual Collection Schema
```
http://localhost:9000/collections/test_json/schema
```

## DuckDB-Specific Features

### Spatial Functions Testing
Test with DuckDB spatial functions if supported in filters:

```
http://localhost:9000/collections/test_json/items?filter=ST_X(geom) > 1.5
```
Expected Response: Points with X coordinate > 1.5

```
http://localhost:9000/collections/test_geom/items?filter=ST_Area(geom) > 0
```
Expected Response: Features with positive area

### Array Operations (if supported in filters)
```
http://localhost:9000/collections/test_arr/items?filter=array_length(val_int) = 3
```
Expected Response: Features where integer array has 3 elements

## Validation Tests

### Geometry Validation
Verify all returned geometries are valid GeoJSON:
- Point coordinates are [x, y] arrays
- LineString coordinates are arrays of [x, y] arrays  
- Polygon coordinates are arrays of linear ring arrays

### Property Validation
Verify returned properties match expected data types:
- Integer properties return as numbers
- Text properties return as strings
- Array properties return as arrays
- JSON properties return as objects/arrays

### Response Format Validation
Verify all responses follow OGC API - Features specification:
- FeatureCollection structure
- Feature structure with type, geometry, properties
- Links array with proper rel values
- Proper HTTP status codes

## Troubleshooting

### Enable Debug Mode
```bash
./duckdb_featureserv --database-path test_spatial.duckdb --debug
```

### Check Server Logs
Look for debug output showing:
- SQL queries being executed
- Geometry conversion details
- Request processing times

### Verify Database Content
```bash
duckdb test_spatial.duckdb -c "LOAD spatial; SELECT table_name, column_name FROM information_schema.columns WHERE data_type = 'GEOMETRY';"
```

### Manual Query Testing
```bash
duckdb test_spatial.duckdb -c "LOAD spatial; SELECT id, ST_AsGeoJSON(geom) FROM test_geom LIMIT 2;"
```
