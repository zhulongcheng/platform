package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/inmem"
	platformtesting "github.com/influxdata/platform/testing"
)

func initUserService(f platformtesting.UserFields, t *testing.T) (platform.UserService, string, func()) {
	t.Helper()
	svc := inmem.NewService()
	svc.IDGenerator = f.IDGenerator

	ctx := context.Background()
	for _, u := range f.Users {
		if err := svc.PutUser(ctx, u); err != nil {
			t.Fatalf("failed to populate users")
		}
	}

	apiBackend := NewMockAPIBackend()
	apiBackend.UserService = svc
	handler := NewUserHandler(apiBackend)
	server := httptest.NewServer(handler)
	client := UserService{
		Addr:     server.URL,
		OpPrefix: inmem.OpPrefix,
	}

	done := server.Close

	return &client, inmem.OpPrefix, done
}

func TestUserService(t *testing.T) {
	t.Parallel()
	platformtesting.UserService(initUserService, t)
}
