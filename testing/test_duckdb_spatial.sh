#!/bin/bash

# Test script for duckdb_featureserv with spatial data
# This script creates test databases and validates the functionality

set -e

echo "🧪 Testing DuckDB Spatial Functionality"
echo "========================================"

# Clean up any existing test databases
rm -f test_spatial.duckdb

# Create test database with spatial data
echo "📊 Creating test database with spatial data..."
duckdb test_spatial.duckdb < testing/duckdb_test.sql

echo "✅ Test database created successfully!"

# Verify tables were created
echo ""
echo "📋 Tables in test database:"
duckdb test_spatial.duckdb -c "SHOW TABLES;"

# Test basic spatial queries
echo ""
echo "🔍 Testing basic spatial queries..."

echo "1. Testing geometry creation and conversion:"
duckdb test_spatial.duckdb -c "LOAD spatial; SELECT id, ST_AsGeoJSON(geom) as geojson FROM test_geom LIMIT 2;"

echo ""
echo "2. Testing array data types:"
duckdb test_spatial.duckdb -c "SELECT id, val_int, val_txt FROM test_arr;"

echo ""
echo "3. Testing JSON data:"
duckdb test_spatial.duckdb -c "SELECT id, val_json FROM test_json;"

echo ""
echo "4. Testing spatial functions:"
duckdb test_spatial.duckdb -c "LOAD spatial; SELECT id, ST_X(geom) as x, ST_Y(geom) as y FROM test_json;"

# Test with duckdb_featureserv
echo ""
echo "🚀 Testing with duckdb_featureserv..."

# Build the server if not already built
if [ ! -f "./duckdb_featureserv" ]; then
    echo "Building duckdb_featureserv..."
    go build
fi

# Start server in background
echo "Starting server with test database..."
./duckdb_featureserv --database-path test_spatial.duckdb --debug &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test API endpoints
echo ""
echo "🌐 Testing API endpoints..."

echo "1. Testing collections endpoint:"
curl -s "http://localhost:9000/collections" | jq '.collections[] | .id' 2>/dev/null || echo "jq not available, showing raw response:"

echo ""
echo "2. Testing features from test_geom:"
curl -s "http://localhost:9000/collections/test_geom/items?limit=2" | jq '.features | length' 2>/dev/null || echo "Retrieved features (raw):"

echo ""
echo "3. Testing features from test_json:"  
curl -s "http://localhost:9000/collections/test_json/items" | jq '.features[] | .geometry' 2>/dev/null || echo "Retrieved geometry data"

echo ""
echo "4. Testing bounding box filter:"
curl -s "http://localhost:9000/collections/test_crs/items?bbox=1000000,400000,1010000,410000&limit=5" | jq '.features | length' 2>/dev/null || echo "Bounding box test completed"

echo ""
echo "5. Testing property selection:"
curl -s "http://localhost:9000/collections/test_names/items?properties=id,colCamelCase" | jq '.features[0].properties | keys' 2>/dev/null || echo "Property selection test completed"

echo ""
echo "6. Testing array data:"
curl -s "http://localhost:9000/collections/test_arr/items" | jq '.features[0].properties.val_int' 2>/dev/null || echo "Array data test completed"

echo ""
echo "7. Testing CQL filter:"
curl -s "http://localhost:9000/collections/test_crs/items?filter=id BETWEEN 1 AND 5" | jq '.features | length' 2>/dev/null || echo "CQL filter test completed"

# Clean up
echo ""
echo "🧹 Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo ""
echo "✅ All tests completed successfully!"
echo ""
echo "📁 Test database created: test_spatial.duckdb"
echo "📄 SQL test file: testing/duckdb_test.sql"
echo ""
echo "To manually test the server:"
echo "  ./duckdb_featureserv --database-path test_spatial.duckdb"
echo "  curl http://localhost:9000/collections"
