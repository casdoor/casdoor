package object

import (
	"database/sql"
	"testing"
	"time"

	"github.com/xorm-io/xorm"
)

func TestGetConnectionPoolSettings(t *testing.T) {
	t.Setenv("dbMaxOpenConns", "40")
	t.Setenv("dbMaxIdleConns", "10")
	t.Setenv("dbConnMaxLifetimeSeconds", "300")

	maxOpenConns, maxIdleConns, maxLifetime := getConnectionPoolSettings()
	if maxOpenConns != 40 {
		t.Errorf("maxOpenConns = %d, want 40", maxOpenConns)
	}
	if maxIdleConns != 10 {
		t.Errorf("maxIdleConns = %d, want 10", maxIdleConns)
	}
	if maxLifetime != 5*time.Minute {
		t.Errorf("maxLifetime = %s, want 5m0s", maxLifetime)
	}
}

func TestConfigureConnectionPool(t *testing.T) {
	t.Setenv("dbMaxOpenConns", "7")

	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	configureConnectionPool(engine)
	if actual := engine.DB().Stats().MaxOpenConnections; actual != 7 {
		t.Errorf("MaxOpenConnections = %d, want 7", actual)
	}
}

func TestNewAdapterFromDbConfiguresConnectionPool(t *testing.T) {
	t.Setenv("dbMaxOpenConns", "7")

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	adapter, err := NewAdapterFromDb("sqlite3", ":memory:", "", db)
	if err != nil {
		t.Fatal(err)
	}
	defer adapter.close()

	if actual := adapter.Engine.DB().Stats().MaxOpenConnections; actual != 7 {
		t.Errorf("MaxOpenConnections = %d, want 7", actual)
	}
}
