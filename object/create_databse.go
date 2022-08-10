package object

import (
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/denisenkom/go-mssqldb/msdsn"
	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// Concatenation of SQL statements may cause SQL injection, but the configuration is performed by the administrator, so it can be ignored

type DBCreator func(dsn, dbName string) (newDSN string, err error)

var adapters = make(map[string]DBCreator)
var reSpace = regexp.MustCompile(`\s+`)

// Register makes a config adapter available by the adapter name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func RegDBCreator(name string, adapter DBCreator) {
	if adapter == nil {
		panic("config: Register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("config: Register called twice for adapter " + name)
	}
	adapters[name] = adapter
}

func QueryDBCreator(driverName string) DBCreator {
	return adapters[driverName]
}

// CreateDatabase create a database based on the information given
func CreateDatabase(driverName, dsn, dbName string) (newDSN string, err error) {
	dbCreator := QueryDBCreator(driverName)
	if dbCreator == nil {
		err = fmt.Errorf("%s are not supported", driverName)
		return
	}
	return dbCreator(dsn, dbName)
}

func postgresDBCreator(dsn, dbName string) (newDSN string, err error) {
	if strings.HasPrefix(dsn, "postgresql://") || strings.HasPrefix(dsn, "postgres://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "", err
		}
		u.Path = dbName
		newDSN = u.String()
		u.Path = ""
		dsn = u.String()
	} else {
		dsnKV := make(map[string]string)
		dsnSplit := reSpace.Split(dsn, -1)
		for _, kvStr := range dsnSplit {
			kv := strings.Split(kvStr, "=")
			if len(kv) != 2 {
				err = fmt.Errorf("error DSN format")
				return
			}
			if kv[0] == "dbname" {
				continue
			}
			dsnKV[kv[0]] = kv[1]
		}
		dsnSplit = make([]string, 0)
		for k, v := range dsnKV {
			dsnSplit = append(dsnSplit, fmt.Sprintf("%s=%s", k, v))
		}
		dsn = strings.Join(dsnSplit, " ")
		newDSN = fmt.Sprintf("%s dbname=%s", dsn, dbName)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		err = fmt.Errorf("failed to connect database: %w", err)
		return
	}
	defer db.Close()
	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			err = nil
		} else {
			err = fmt.Errorf("failed to create database: %w", err)
		}
	}
	return
}

func mysqlDBCreator(dsn, dbName string) (newDSN string, err error) {
	var cfg *mysql.Config
	cfg, err = mysql.ParseDSN(dsn)
	if err != nil {
		err = fmt.Errorf("error DSN format: %w", err)
		return
	}
	cfg.DBName = ""
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		err = fmt.Errorf("failed to connect database: %w", err)
		return
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s default charset utf8mb4 COLLATE utf8mb4_general_ci", dbName))
	if err != nil {
		err = fmt.Errorf("failed to create database: %w", err)
	}
	cfg.DBName = dbName
	newDSN = cfg.FormatDSN()
	return
}

func sqlite3DBCreator(dsn, dbName string) (newDSN string, err error) {
	return dsn, nil
}

func mssqlDBCreator(dsn, dbName string) (newDSN string, err error) {
	cfg, _, err := msdsn.Parse(dsn)
	if err != nil {
		err = fmt.Errorf("error DSN format: %w", err)
		return
	}
	cfg.Database = ""
	dsn = cfg.URL().String()
	db, err := sql.Open("mssql", cfg.URL().String())
	if err != nil {
		err = fmt.Errorf("failed to connect database: %w", err)
		return
	}
	defer db.Close()
	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			err = nil
		} else {
			err = fmt.Errorf("failed to create database: %w", err)
		}
	}
	cfg.Database = dbName
	newDSN = cfg.URL().String()
	return
}

func init() {
	RegDBCreator("postgres", postgresDBCreator)
	RegDBCreator("mysql", mysqlDBCreator)
	RegDBCreator("sqlite3", sqlite3DBCreator)
	RegDBCreator("mssql", mssqlDBCreator)
}
