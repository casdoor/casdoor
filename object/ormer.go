// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"

	"github.com/beego/beego/v2/server/web"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	_ "github.com/denisenkom/go-mssqldb" // db = mssql
	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq" // db = postgres
	"github.com/xorm-io/xorm"
	"github.com/xorm-io/xorm/core"
	"github.com/xorm-io/xorm/names"
	_ "modernc.org/sqlite" // db = sqlite
)

const (
	defaultConfigPath     = "conf/app.conf"
	defaultExportFilePath = "init_data_dump.json"
	mysqlTLSConfigName    = "custom-mtls"
)

var (
	ormer          *Ormer = nil
	createDatabase        = true
	configPath            = defaultConfigPath
	exportData            = false
	exportFilePath        = defaultExportFilePath
)

func InitFlag() {
	createDatabasePtr := flag.Bool("createDatabase", false, "true if you need to create database")
	configPathPtr := flag.String("config", defaultConfigPath, "set it to \"/your/path/app.conf\" if your config file is not in: \"/conf/app.conf\"")
	exportDataPtr := flag.Bool("export", false, "export database to JSON file and exit (use -exportPath to specify custom location)")
	exportFilePathPtr := flag.String("exportPath", defaultExportFilePath, "path to the exported data file (used with -export)")
	flag.Parse()

	createDatabase = *createDatabasePtr
	configPath = *configPathPtr
	exportData = *exportDataPtr
	exportFilePath = *exportFilePathPtr
}

func ShouldExportData() bool {
	return exportData
}

func GetExportFilePath() string {
	return exportFilePath
}

// setupMySQLTLS configures TLS for MySQL connections if certificate paths are provided
func setupMySQLTLS() error {
	caCertPath := conf.GetConfigString("dbCaCert")
	clientCertPath := conf.GetConfigString("dbClientCert")
	clientKeyPath := conf.GetConfigString("dbClientKey")

	// If no certificates are configured, return nil (no TLS)
	if caCertPath == "" && clientCertPath == "" && clientKeyPath == "" {
		return nil
	}

	// Create TLS config
	tlsConfig := &tls.Config{}

	// Load CA certificate if provided
	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return fmt.Errorf("failed to read CA certificate from %s: %w", caCertPath, err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("failed to parse CA certificate from %s", caCertPath)
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if both are provided
	if clientCertPath != "" && clientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			return fmt.Errorf("failed to load client certificate/key from %s and %s: %w", clientCertPath, clientKeyPath, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	} else if clientCertPath != "" || clientKeyPath != "" {
		// If only one is provided, return an error
		return fmt.Errorf("both dbClientCert and dbClientKey must be provided together")
	}

	// Register the TLS config with MySQL driver
	err := mysql.RegisterTLSConfig(mysqlTLSConfigName, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to register MySQL TLS config: %w", err)
	}

	return nil
}

// isMySQLTLSConfigured returns true if any MySQL TLS certificate is configured
func isMySQLTLSConfigured() bool {
	caCertPath := conf.GetConfigString("dbCaCert")
	clientCertPath := conf.GetConfigString("dbClientCert")
	clientKeyPath := conf.GetConfigString("dbClientKey")
	return caCertPath != "" || clientCertPath != "" || clientKeyPath != ""
}

// appendMySQLTLSParam appends the TLS parameter to MySQL DSN if TLS is configured
func appendMySQLTLSParam(dsn string) string {
	// If no certificates are configured, return the original DSN
	if !isMySQLTLSConfigured() {
		return dsn
	}

	// Append the TLS parameter
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	return dsn + separator + "tls=" + mysqlTLSConfigName
}

func InitConfig() {
	err := web.LoadAppConfig("ini", "../conf/app.conf")
	if err != nil {
		panic(err)
	}

	web.BConfig.WebConfig.Session.SessionOn = true

	InitAdapter()
	CreateTables()
}

func InitAdapter() {
	if conf.GetConfigString("driverName") == "" {
		if !util.FileExist(configPath) {
			dir, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			dir = strings.ReplaceAll(dir, "\\", "/")
			panic(fmt.Sprintf("The Casdoor config file: \"app.conf\" was not found, it should be placed at: \"%s/conf/app.conf\"", dir))
		}
	}

	// Setup MySQL TLS if certificates are configured
	if conf.GetConfigString("driverName") == "mysql" {
		err := setupMySQLTLS()
		if err != nil {
			panic(err)
		}
	}

	if createDatabase {
		err := createDatabaseForPostgres(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
		if err != nil {
			panic(err)
		}
	}

	var err error
	ormer, err = NewAdapter(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
	if err != nil {
		panic(err)
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tbMapper := names.NewPrefixMapper(names.SnakeMapper{}, tableNamePrefix)
	ormer.Engine.SetTableMapper(tbMapper)
}

func CreateTables() {
	if createDatabase {
		err := ormer.CreateDatabase()
		if err != nil {
			panic(err)
		}
	}

	ormer.createTable()
}

// Ormer represents the MySQL adapter for policy storage.
type Ormer struct {
	driverName     string
	dataSourceName string
	dbName         string
	Db             *sql.DB
	Engine         *xorm.Engine
}

// finalizer is the destructor for Ormer.
func finalizer(a *Ormer) {
	err := a.Engine.Close()
	if err != nil {
		panic(err)
	}

	if a.Db != nil {
		err = a.Db.Close()
		if err != nil {
			panic(err)
		}
	}
}

// NewAdapter is the constructor for Ormer.
func NewAdapter(driverName string, dataSourceName string, dbName string) (*Ormer, error) {
	a := &Ormer{}
	a.driverName = driverName
	a.dataSourceName = dataSourceName
	a.dbName = dbName

	// Open the DB, create it if not existed.
	err := a.open()
	if err != nil {
		return nil, err
	}

	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, finalizer)

	return a, nil
}

// NewAdapterFromDb is the constructor for Ormer.
func NewAdapterFromDb(driverName string, dataSourceName string, dbName string, db *sql.DB) (*Ormer, error) {
	a := &Ormer{}
	a.driverName = driverName
	a.dataSourceName = dataSourceName
	a.dbName = dbName
	a.Db = db

	// Open the DB, create it if not existed.
	err := a.openFromDb(a.Db)
	if err != nil {
		return nil, err
	}

	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, finalizer)

	return a, nil
}

func refineDataSourceNameForPostgres(dataSourceName string) string {
	reg := regexp.MustCompile(`dbname=[^ ]+\s*`)
	return reg.ReplaceAllString(dataSourceName, "dbname=postgres")
}

func createDatabaseForPostgres(driverName string, dataSourceName string, dbName string) error {
	if driverName == "postgres" {
		db, err := sql.Open(driverName, refineDataSourceNameForPostgres(dataSourceName))
		if err != nil {
			return err
		}
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\";", dbName))
		if err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
		schema := util.GetValueFromDataSourceName("search_path", dataSourceName)
		if schema != "" {
			db, err = sql.Open(driverName, dataSourceName)
			if err != nil {
				return err
			}
			defer db.Close()

			_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %s;", schema))
			if err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return err
				}
			}
		}

		return nil
	} else {
		return nil
	}
}

func (a *Ormer) CreateDatabase() error {
	if a.driverName == "postgres" {
		return nil
	}

	dataSourceName := a.dataSourceName
	if a.driverName == "mysql" {
		dataSourceName = appendMySQLTLSParam(dataSourceName)
	}

	engine, err := xorm.NewEngine(a.driverName, dataSourceName)
	if err != nil {
		return err
	}
	defer engine.Close()

	_, err = engine.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s default charset utf8mb4 COLLATE utf8mb4_general_ci", a.dbName))
	return err
}

func (a *Ormer) open() error {
	dataSourceName := a.dataSourceName + a.dbName
	if a.driverName != "mysql" {
		dataSourceName = a.dataSourceName
	} else {
		dataSourceName = appendMySQLTLSParam(dataSourceName)
	}

	engine, err := xorm.NewEngine(a.driverName, dataSourceName)
	if err != nil {
		return err
	}

	if a.driverName == "postgres" {
		schema := util.GetValueFromDataSourceName("search_path", dataSourceName)
		if schema != "" {
			engine.SetSchema(schema)
		}
	}

	a.Engine = engine
	return nil
}

func (a *Ormer) openFromDb(db *sql.DB) error {
	dataSourceName := a.dataSourceName + a.dbName
	if a.driverName != "mysql" {
		dataSourceName = a.dataSourceName
	} else {
		dataSourceName = appendMySQLTLSParam(dataSourceName)
	}

	xormDb := core.FromDB(db)

	engine, err := xorm.NewEngineWithDB(a.driverName, dataSourceName, xormDb)
	if err != nil {
		return err
	}

	if a.driverName == "postgres" {
		schema := util.GetValueFromDataSourceName("search_path", dataSourceName)
		if schema != "" {
			engine.SetSchema(schema)
		}
	}

	a.Engine = engine
	return nil
}

func (a *Ormer) close() {
	_ = a.Engine.Close()
	a.Engine = nil
}

func (a *Ormer) createTable() {
	showSql := conf.GetConfigBool("showSql")
	a.Engine.ShowSQL(showSql)

	err := a.Engine.Sync2(new(Organization))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Group))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(User))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Invitation))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Application))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Provider))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Resource))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Cert))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Role))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Permission))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Model))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Adapter))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Enforcer))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Session))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Token))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Product))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Payment))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Order))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Plan))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Pricing))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Subscription))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Transaction))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Syncer))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(casvisorsdk.Record))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Webhook))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(VerificationRecord))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Ldap))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(RadiusAccounting))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(xormadapter.CasbinRule))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Form))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Ticket))
	if err != nil {
		panic(err)
	}
}
