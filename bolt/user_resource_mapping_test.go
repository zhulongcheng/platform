package bolt_test

import (
	"context"
	"testing"

	"github.com/influxdata/platform"
	platformtesting "github.com/influxdata/platform/testing"
)

func initUserResourceMappingService(f platformtesting.UserResourceFields, t *testing.T) (platform.UserResourceMappingService, func()) {
	c, closeFn, err := NewTestClient()
	if err != nil {
		t.Fatalf("failed to create new bolt client: %v", err)
	}
	ctx := context.Background()
	for _, m := range f.UserResourceMappings {
		switch m.Resource {
		case platform.BucketsResource:
			if err := c.PutOrganization(ctx, &platform.Organization{ID: m.ResourceID, Name: m.ResourceID.String()}); err != nil {
				t.Fatalf("failed to populate orgs")
			}
			if err := c.PutBucket(ctx, &platform.Bucket{ID: m.ResourceID, OrganizationID: m.ResourceID}); err != nil {
				t.Fatalf("failed to populate buckets")
			}
		case platform.DashboardsResource:
			if err := c.PutDashboard(ctx, &platform.Dashboard{ID: m.ResourceID, Name: m.ResourceID.String()}); err != nil {
				t.Fatalf("failed to populate dashboards")
			}
		}

		if err := c.CreateUserResourceMapping(ctx, m); err != nil {
			t.Fatalf("failed to populate mappings")
		}
	}

	return c, func() {
		defer closeFn()
		for _, m := range f.UserResourceMappings {
			if err := c.DeleteUserResourceMapping(ctx, m.ResourceID, m.UserID); err != nil {
				t.Logf("failed to remove user resource mapping: %v", err)
			}
		}
	}
}

func TestUserResourceMappingService_FindUserResourceMappings(t *testing.T) {
	platformtesting.FindUserResourceMappings(initUserResourceMappingService, t)
}

func TestUserResourceMappingService_CreateUserResourceMapping(t *testing.T) {
	platformtesting.CreateUserResourceMapping(initUserResourceMappingService, t)
}

func TestUserResourceMappingService_DeleteUserResourceMapping(t *testing.T) {
	platformtesting.DeleteUserResourceMapping(initUserResourceMappingService, t)
}
