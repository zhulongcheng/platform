package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/platform/mock"
	"go.uber.org/zap"
)

func TestAPIHandler_NotFound(t *testing.T) {
	type args struct {
		method string
		path   string
	}
	type wants struct {
		statusCode  int
		contentType string
		body        string
	}

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "path not found",
			args: args{
				method: "GET",
				path:   "/404",
			},
			wants: wants{
				statusCode:  http.StatusNotFound,
				contentType: "application/json; charset=utf-8",
				body: `
{
  "code": "not found",
  "message": "path not found"
}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			w := httptest.NewRecorder()

			b := &APIBackend{}
			b.Logger = zap.NewNop()

			h := NewAPIHandler(b)
			h.ServeHTTP(w, r)

			res := w.Result()
			content := res.Header.Get("Content-Type")
			body, _ := ioutil.ReadAll(res.Body)

			if res.StatusCode != tt.wants.statusCode {
				t.Errorf("%q. get %v, want %v", tt.name, res.StatusCode, tt.wants.statusCode)
			}
			if tt.wants.contentType != "" && content != tt.wants.contentType {
				t.Errorf("%q. get %v, want %v", tt.name, content, tt.wants.contentType)
			}
			if eq, _ := jsonEqual(string(body), tt.wants.body); tt.wants.body != "" && !eq {
				t.Errorf("%q. get\n***%v***\n,\nwant\n***%v***", tt.name, string(body), tt.wants.body)
			}

		})
	}
}

// NewMockAPIBackend returns a APIBackend with mock services.
func NewMockAPIBackend() *APIBackend {
	return &APIBackend{
		DeveloperMode: true,
		Logger:        zap.NewNop(),

		AuthorizationService:       mock.NewAuthorizationService(),
		BucketService:              mock.NewBucketService(),
		SessionService:             mock.NewSessionService(),
		UserService:                mock.NewUserService(),
		OrganizationService:        &mock.OrganizationService{},
		UserResourceMappingService: mock.NewUserResourceMappingService(),
		LabelService:               mock.NewLabelService(),
		DashboardService:           &mock.DashboardService{},
		ViewService:                &mock.ViewService{},
		MacroService:               &mock.MacroService{},
		BasicAuthService:           mock.NewBasicAuthService("", ""),
		TaskService:                &mock.TaskService{},
		SecretService:              mock.NewSecretService(),

		// TODO impl mock service
		//SourceService:                   mock.NewSourceService(),
		//OnboardingService:               mock.NewOnboardingService(),
		//ProxyQueryService:               mock.NewProxyQueryService(),
		//PointsWriter:                    mock.NewPointsWriter(),
		//DashboardOperationLogService:    mock.NewDashboardOperationLogService(),
		//BucketOperationLogService:       mock.NewBucketOperationLogService{},
		//UserOperationLogService:         mock.NewUserOperationLogService(),
		//OrganizationOperationLogService: mock.NewOrganizationOperationLogService(),
		//TelegrafService:                 mock.NewTelegrafService(),
		//ScraperTargetStoreService:       mock.NewScraperTargetStoreService(),
	}
}
