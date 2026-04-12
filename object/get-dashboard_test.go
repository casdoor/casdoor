package object

import (
	"sync"
	"testing"
)

func TestGetDashboardMapItemUsesPrefixedTableName(t *testing.T) {
	var dashboardMap sync.Map
	expected := DashboardMapItem{itemCount: 3}
	dashboardMap.Store("casdoor_user", expected)

	got, ok := getDashboardMapItem(&dashboardMap, "casdoor_", "user")
	if !ok {
		t.Fatalf("expected prefixed dashboard item to be found")
	}
	if got.itemCount != expected.itemCount {
		t.Fatalf("expected itemCount=%d, got %d", expected.itemCount, got.itemCount)
	}
}

func TestGetDashboardMapItemSupportsEmptyPrefix(t *testing.T) {
	var dashboardMap sync.Map
	expected := DashboardMapItem{itemCount: 5}
	dashboardMap.Store("user", expected)

	got, ok := getDashboardMapItem(&dashboardMap, "", "user")
	if !ok {
		t.Fatalf("expected unprefixed dashboard item to be found")
	}
	if got.itemCount != expected.itemCount {
		t.Fatalf("expected itemCount=%d, got %d", expected.itemCount, got.itemCount)
	}
}
