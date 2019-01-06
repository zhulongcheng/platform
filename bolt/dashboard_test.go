package bolt_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/bolt"
	platformtesting "github.com/influxdata/platform/testing"
)

func initDashboardService(f platformtesting.DashboardFields, t *testing.T) (platform.DashboardService, string, func()) {
	c, closeFn, err := NewTestClient()
	if err != nil {
		t.Fatalf("failed to create new bolt client: %v", err)
	}
	c.IDGenerator = f.IDGenerator

	if f.NowFn == nil {
		f.NowFn = time.Now
	}
	c.WithTime(f.NowFn)
	ctx := context.TODO()
	for _, b := range f.Dashboards {
		if err := c.PutDashboard(ctx, b); err != nil {
			t.Fatalf("failed to populate dashboards")
		}
	}
	for _, b := range f.Views {
		if err := c.PutView(ctx, b); err != nil {
			t.Fatalf("failed to populate views")
		}
	}
	return c, bolt.OpPrefix, func() {
		defer closeFn()
		for _, b := range f.Dashboards {
			if err := c.DeleteDashboard(ctx, b.ID); err != nil {
				t.Logf("failed to remove dashboard: %v", err)
			}
		}
		for _, b := range f.Views {
			if err := c.DeleteView(ctx, b.ID); err != nil {
				t.Logf("failed to remove view: %v", err)
			}
		}
	}
}

func TestDashboardService(t *testing.T) {
	platformtesting.DashboardService(initDashboardService, t)
}
