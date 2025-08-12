--===============================================
-- DDL for database objects to test duckdb_featureserv
-- Translated from PostgreSQL/PostGIS to DuckDB/duckdb-spatial
--===============================================

-- Install and load spatial extension
INSTALL spatial;
LOAD spatial;

-- Note: DuckDB doesn't have schemas in the same way as PostgreSQL
-- We'll use table name prefixes instead

--=====================================================================
-- Test table with specific CRS (EPSG:3005)

CREATE TABLE test_crs (
    id INTEGER PRIMARY KEY,
    geom GEOMETRY,
    name VARCHAR
);

-- DROP TABLE test_crs;
-- DELETE FROM test_crs;

-- DuckDB equivalent using generate_series and ST_MakeEnvelope
INSERT INTO test_crs
SELECT ROW_NUMBER() OVER () AS id,
       ST_GeomFromText(
           'POLYGON((' || 
           (1000000.0 + 20000 * x) || ' ' || (400000.0 + 20000 * y) || ', ' ||
           (1000000.0 + 20000 * (x + 1)) || ' ' || (400000.0 + 20000 * y) || ', ' ||
           (1000000.0 + 20000 * (x + 1)) || ' ' || (400000.0 + 20000 * (y + 1)) || ', ' ||
           (1000000.0 + 20000 * x) || ' ' || (400000.0 + 20000 * (y + 1)) || ', ' ||
           (1000000.0 + 20000 * x) || ' ' || (400000.0 + 20000 * y) || '))'
       ) AS geom,
       CAST(x AS VARCHAR) || '_' || CAST(y AS VARCHAR) AS name
FROM generate_series(0, 9) AS t1(x)
CROSS JOIN generate_series(0, 9) AS t2(y);

--=====================================================================
-- Test table with SRID 0 (no specific projection)

CREATE TABLE test_srid0 (
    id INTEGER PRIMARY KEY,
    geom GEOMETRY,
    name VARCHAR
);

-- DROP TABLE test_srid0;
-- DELETE FROM test_srid0;

INSERT INTO test_srid0
SELECT ROW_NUMBER() OVER () AS id,
       ST_GeomFromText(
           'POLYGON((' || 
           (1.0 + 2 * x) || ' ' || (4.0 + 2 * y) || ', ' ||
           (1.0 + 2 * (x + 1)) || ' ' || (4.0 + 2 * y) || ', ' ||
           (1.0 + 2 * (x + 1)) || ' ' || (4.0 + 2 * (y + 1)) || ', ' ||
           (1.0 + 2 * x) || ' ' || (4.0 + 2 * (y + 1)) || ', ' ||
           (1.0 + 2 * x) || ' ' || (4.0 + 2 * y) || '))'
       ) AS geom,
       CAST(x AS VARCHAR) || '_' || CAST(y AS VARCHAR) AS name
FROM generate_series(0, 9) AS t1(x)
CROSS JOIN generate_series(0, 9) AS t2(y);

--=====================================================================
-- Test table with JSON data

CREATE TABLE test_json (
    id INTEGER PRIMARY KEY,
    geom GEOMETRY,
    val_json JSON
);

-- DROP TABLE test_json;

INSERT INTO test_json
VALUES
  (1, ST_GeomFromText('POINT(1 1)'), '["a", "b", "c"]'),
  (2, ST_GeomFromText('POINT(2 2)'), '{"p1": 1, "p2": 2.3, "p3": [1, 2, 3]}');

--=====================================================================
-- Test a table with mixed geometry types

CREATE TABLE test_geom (
    id INTEGER PRIMARY KEY,
    geom GEOMETRY,
    data VARCHAR
);

-- DROP TABLE test_geom;

INSERT INTO test_geom
VALUES
  (1, ST_GeomFromText('POINT(1 1)'), 'aaa'),
  (2, ST_GeomFromText('LINESTRING(1 1, 2 2)'), 'bbb');

--=====================================================================
-- Test array handling
-- Note: DuckDB uses array syntax differently than PostgreSQL

CREATE TABLE test_arr (
    id INTEGER PRIMARY KEY,
    geom GEOMETRY,
    val_bool BOOLEAN[],
    val_int INTEGER[],
    val_dp DOUBLE[],
    val_txt VARCHAR[]
);

-- DROP TABLE test_arr;

INSERT INTO test_arr
VALUES (1, ST_GeomFromText('POINT(1 1)'),
        [true, true, false],
        [1, 2, 3],
        [1.1, 2.2, 3.3],
        ['a', 'bb', 'ccc']);

--=====================================================================
-- Test column name handling (quoted identifiers)

CREATE TABLE test_names (
    id INTEGER PRIMARY KEY,
    geom GEOMETRY,
    "colCamelCase" INTEGER
);

-- DROP TABLE test_names;

INSERT INTO test_names
VALUES 
    (1, ST_GeomFromText('POINT(1 1)'), 1),
    (2, ST_GeomFromText('POINT(2 2)'), 2);

--=====================================================================
-- Test functions
-- Note: DuckDB doesn't have the same function creation syntax as PostgreSQL
-- This is a simplified example showing how spatial functions work

-- Example queries demonstrating DuckDB spatial functions:

-- Calculate area of geometries
SELECT id, name, ST_Area(geom) as area 
FROM test_crs 
LIMIT 5;

-- Convert geometries to GeoJSON
SELECT id, data, ST_AsGeoJSON(geom) as geojson 
FROM test_geom;

-- Buffer operation
SELECT id, data, ST_AsGeoJSON(ST_Buffer(geom, 0.1)) as buffered_geom
FROM test_geom;

-- Distance calculation
SELECT a.id as id1, b.id as id2, 
       ST_Distance(a.geom, b.geom) as distance
FROM test_geom a, test_geom b 
WHERE a.id < b.id;

-- Spatial intersection check
SELECT COUNT(*) as intersecting_pairs
FROM test_crs a, test_crs b 
WHERE a.id < b.id 
  AND ST_Intersects(a.geom, b.geom);

--=====================================================================
-- Test spatial indexing and performance
-- Note: DuckDB spatial doesn't have the same indexing as PostGIS yet

-- Bounding box queries
SELECT id, name 
FROM test_crs 
WHERE ST_Intersects(geom, ST_GeomFromText('POLYGON((1000000 400000, 1040000 400000, 1040000 440000, 1000000 440000, 1000000 400000))'));

-- Point-in-polygon queries
SELECT COUNT(*) as points_in_first_polygon
FROM test_json
WHERE ST_Within(geom, (SELECT geom FROM test_crs WHERE id = 1));

--=====================================================================
-- Additional DuckDB-specific spatial features

-- Extract coordinates from points
SELECT id, 
       ST_X(geom) as longitude,
       ST_Y(geom) as latitude
FROM test_json;

-- Geometry type information
SELECT id, 
       ST_GeometryType(geom) as geom_type
FROM test_geom;

-- Create geometries from coordinates
SELECT 1 as id, 
       ST_Point(13.4050, 52.5200) as berlin_location,
       'Berlin' as city;

--=====================================================================
-- Data validation and error handling

-- Check for valid geometries
SELECT id, ST_IsValid(geom) as is_valid
FROM test_geom;

-- Get geometry bounds (using aggregation)
SELECT 'test_crs' as table_name,
       MIN(ST_XMin(geom)) as min_x,
       MIN(ST_YMin(geom)) as min_y,
       MAX(ST_XMax(geom)) as max_x,
       MAX(ST_YMax(geom)) as max_y
FROM test_crs;
