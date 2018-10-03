package http

// import (
// 	"context"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/influxdata/platform/inmem"
// 	platformtesting "github.com/influxdata/platform/testing"
// )

// func initOnboardingService(f platformtesting.OnboardingFields, t *testing.T) (platformtesting.OnBoardingNBasicAuthService, func()) {
// 	t.Helper()
// 	svc := inmem.NewService()
// 	svc.IDGenerator = f.IDGenerator

// 	ctx := context.Background()
// 	if err := svc.PutOnboardingStatus(ctx, !f.IsOnboarding); err != nil {
// 		t.Fatalf("failed to set new onboarding finished: %v", err)
// 	}

// 	handler := NewSetupHandler()
// 	handler.OnboardingService = svc
// 	server := httptest.NewServer(handler)
// 	client := Service{
// 		Addr: server.URL,
// 	}
// 	done := server.Close

// 	return &client, done
// }
// func TestOnboardingService(t *testing.T) {
// 	t.Parallel()
// 	platformtesting.Generate(initOnboardingService, t)
// }
