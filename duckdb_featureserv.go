package main

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

/*
# Running
Usage: ./duckdb_featureserv [ -test ] [ --database-path /path/to/database.db ]

Browser: e.g. http://localhost:9000/index.html

# Configuration
DuckDB file path in env var `DUCKDBFS_DATABASE_PATH`
Example: `export DUCKDBFS_DATABASE_PATH="/path/to/database.db"`

Table filtering via env vars `DUCKDBFS_DATABASE_TABLEINCLUDES` and `DUCKDBFS_DATABASE_TABLEEXCLUDES` (optional)
Examples:
  `export DUCKDBFS_DATABASE_TABLEINCLUDES="public,reports.monthly"`
  `export DUCKDBFS_DATABASE_TABLEEXCLUDES="private,system"`
If not specified, all tables with geometry columns will be served as collections

For backward compatibility, the old environment variable `DUCKDB_PATH` is still supported but deprecated.

# Logging
Logging to stdout
*/

import (
	"fmt"
	"os"

	"github.com/tobilg/duckdb_featureserv/internal/conf"
	"github.com/tobilg/duckdb_featureserv/internal/data"
	"github.com/tobilg/duckdb_featureserv/internal/service"
	"github.com/tobilg/duckdb_featureserv/internal/ui"

	"github.com/pborman/getopt/v2"
	log "github.com/sirupsen/logrus"
)

var flagTestModeOn bool
var flagDebugOn bool
var flagDevModeOn bool
var flagHelp bool
var flagVersion bool
var flagConfigFilename string
var flagDuckDBPath string

var flagDisableUi bool

func init() {
	initCommnandOptions()
}

func initCommnandOptions() {
	getopt.FlagLong(&flagHelp, "help", '?', "Show command usage")
	getopt.FlagLong(&flagConfigFilename, "config", 'c', "", "config file name")
	getopt.FlagLong(&flagDebugOn, "debug", 'd', "Set logging level to TRACE")
	getopt.FlagLong(&flagDevModeOn, "devel", 0, "Run in development mode")
	getopt.FlagLong(&flagTestModeOn, "test", 't', "Serve mock data for testing")
	getopt.FlagLong(&flagVersion, "version", 'v', "Output the version information")
	getopt.FlagLong(&flagDuckDBPath, "database-path", 0, "", "Path to DuckDB database file")

	getopt.FlagLong(&flagDisableUi, "disable-ui", 0, "Disable HTML UI routes")
}

func main() {
	getopt.Parse()

	if flagHelp {
		getopt.Usage()
		os.Exit(1)
	}

	if flagVersion {
		fmt.Printf("%s %s\n", conf.AppConfig.Name, conf.AppConfig.Version)
		os.Exit(1)
	}

	log.Infof("----  %s - Version %s ----------\n", conf.AppConfig.Name, conf.AppConfig.Version)

	conf.InitConfig(flagConfigFilename, flagDebugOn)

	// Set DuckDB parameters from command line if provided
	if flagDuckDBPath != "" {
		conf.Configuration.Database.DatabasePath = flagDuckDBPath
	}

	// Set UI disable flag from command line
	if flagDisableUi {
		conf.Configuration.Server.DisableUi = true
	}

	if flagTestModeOn || flagDevModeOn {
		ui.HTMLDynamicLoad = true
		log.Info("Running in development mode")
	}
	// Commandline over-rides config file for debugging
	if flagDebugOn || conf.Configuration.Server.Debug {
		log.SetLevel(log.TraceLevel)
		log.Debugf("Log level = DEBUG\n")
	}
	conf.DumpConfig()

	//-- Initialize catalog (with DB conn if used)
	var catalog data.Catalog
	if flagTestModeOn {
		catalog = data.CatMockInstance()
	} else {
		catalog = data.CatDBInstance()
	}
	includes := conf.Configuration.Database.TableIncludes
	excludes := conf.Configuration.Database.TableExcludes
	catalog.SetIncludeExclude(includes, excludes)

	//-- Start up service
	service.Initialize()
	service.Serve(catalog)
}
