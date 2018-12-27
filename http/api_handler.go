package http

import (
	http "net/http"
	"strings"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/chronograf/server"
	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/storage"
	"go.uber.org/zap"
)

// APIHandler is a collection of all the service handlers.
type APIHandler struct {
	BucketHandler        *BucketHandler
	UserHandler          *UserHandler
	OrgHandler           *OrgHandler
	AuthorizationHandler *AuthorizationHandler
	DashboardHandler     *DashboardHandler
	AssetHandler         *AssetHandler
	ChronografHandler    *ChronografHandler
	ViewHandler          *ViewHandler
	SourceHandler        *SourceHandler
	MacroHandler         *MacroHandler
	TaskHandler          *TaskHandler
	TelegrafHandler      *TelegrafHandler
	QueryHandler         *FluxHandler
	WriteHandler         *WriteHandler
	SetupHandler         *SetupHandler
	SessionHandler       *SessionHandler
}

// APIBackend is all services and associated parameters required to construct
// an APIHandler.
type APIBackend struct {
	DeveloperMode bool
	Logger        *zap.Logger

	NewBucketService func(*platform.Source) (platform.BucketService, error)
	NewQueryService  func(*platform.Source) (query.ProxyQueryService, error)

	PointsWriter                    storage.PointsWriter
	AuthorizationService            platform.AuthorizationService
	BucketService                   platform.BucketService
	SessionService                  platform.SessionService
	UserService                     platform.UserService
	OrganizationService             platform.OrganizationService
	UserResourceMappingService      platform.UserResourceMappingService
	LabelService                    platform.LabelService
	DashboardService                platform.DashboardService
	DashboardOperationLogService    platform.DashboardOperationLogService
	BucketOperationLogService       platform.BucketOperationLogService
	UserOperationLogService         platform.UserOperationLogService
	OrganizationOperationLogService platform.OrganizationOperationLogService
	ViewService                     platform.ViewService
	SourceService                   platform.SourceService
	MacroService                    platform.MacroService
	BasicAuthService                platform.BasicAuthService
	OnboardingService               platform.OnboardingService
	ProxyQueryService               query.ProxyQueryService
	TaskService                     platform.TaskService
	TelegrafService                 platform.TelegrafConfigStore
	ScraperTargetStoreService       platform.ScraperTargetStoreService
	ChronografService               *server.Service
}

// NewAPIHandler constructs all api handlers beneath it and returns an APIHandler
func NewAPIHandler(b *APIBackend) *APIHandler {
	h := &APIHandler{}
	h.SessionHandler = NewSessionHandler()
	h.SessionHandler.BasicAuthService = b.BasicAuthService
	h.SessionHandler.SessionService = b.SessionService
	h.SessionHandler.Logger = b.Logger.With(zap.String("handler", "basicAuth"))

	h.BucketHandler = NewBucketHandler(b)

	h.OrgHandler = NewOrgHandler(b.UserResourceMappingService, b.LabelService, b.UserService)
	h.OrgHandler.OrganizationService = b.OrganizationService
	h.OrgHandler.BucketService = b.BucketService
	h.OrgHandler.OrganizationOperationLogService = b.OrganizationOperationLogService

	h.UserHandler = NewUserHandler()
	h.UserHandler.UserService = b.UserService
	h.UserHandler.BasicAuthService = b.BasicAuthService
	h.UserHandler.UserOperationLogService = b.UserOperationLogService

	h.DashboardHandler = NewDashboardHandler(b.UserResourceMappingService, b.LabelService, b.UserService)
	h.DashboardHandler.DashboardService = b.DashboardService
	h.DashboardHandler.DashboardOperationLogService = b.DashboardOperationLogService

	h.ViewHandler = NewViewHandler(b.UserResourceMappingService, b.LabelService, b.UserService)
	h.ViewHandler.ViewService = b.ViewService

	h.MacroHandler = NewMacroHandler()
	h.MacroHandler.MacroService = b.MacroService

	h.AuthorizationHandler = NewAuthorizationHandler(b.UserService)
	h.AuthorizationHandler.AuthorizationService = b.AuthorizationService
	h.AuthorizationHandler.Logger = b.Logger.With(zap.String("handler", "auth"))

	h.SourceHandler = NewSourceHandler()
	h.SourceHandler.SourceService = b.SourceService
	h.SourceHandler.NewBucketService = b.NewBucketService
	h.SourceHandler.NewQueryService = b.NewQueryService

	h.SetupHandler = NewSetupHandler()
	h.SetupHandler.OnboardingService = b.OnboardingService

	h.TaskHandler = NewTaskHandler(b.UserResourceMappingService, b.LabelService, b.Logger, b.UserService)
	h.TaskHandler.TaskService = b.TaskService
	h.TaskHandler.AuthorizationService = b.AuthorizationService
	h.TaskHandler.UserResourceMappingService = b.UserResourceMappingService

	h.TelegrafHandler = NewTelegrafHandler(
		b.Logger.With(zap.String("handler", "telegraf")),
		b.UserResourceMappingService,
		b.LabelService,
		b.TelegrafService,
		b.UserService,
	)

	h.WriteHandler = NewWriteHandler(b.PointsWriter)
	h.WriteHandler.OrganizationService = b.OrganizationService
	h.WriteHandler.BucketService = b.BucketService
	h.WriteHandler.Logger = b.Logger.With(zap.String("handler", "write"))

	h.QueryHandler = NewFluxHandler()
	h.QueryHandler.OrganizationService = b.OrganizationService
	h.QueryHandler.Logger = b.Logger.With(zap.String("handler", "query"))
	h.QueryHandler.ProxyQueryService = b.ProxyQueryService

	h.ChronografHandler = NewChronografHandler(b.ChronografService)

	return h
}

var apiLinks = map[string]interface{}{
	// when adding new links, please take care to keep this list alphabetical
	// as this makes it easier to verify values against the swagger document.
	"authorizations": "/api/v2/authorizations",
	"buckets":        "/api/v2/buckets",
	"dashboards":     "/api/v2/dashboards",
	"external": map[string]string{
		"statusFeed": "https://www.influxdata.com/feed/json",
	},
	"macros": "/api/v2/macros",
	"me":     "/api/v2/me",
	"orgs":   "/api/v2/orgs",
	"query": map[string]string{
		"self":        "/api/v2/query",
		"ast":         "/api/v2/query/ast",
		"analyze":     "/api/v2/query/analyze",
		"spec":        "/api/v2/query/spec",
		"suggestions": "/api/v2/query/suggestions",
	},
	"setup":   "/api/v2/setup",
	"signin":  "/api/v2/signin",
	"signout": "/api/v2/signout",
	"sources": "/api/v2/sources",
	"system": map[string]string{
		"metrics": "/metrics",
		"debug":   "/debug/pprof",
		"health":  "/health",
	},
	"tasks":     "/api/v2/tasks",
	"telegrafs": "/api/v2/telegrafs",
	"users":     "/api/v2/users",
	"views":     "/api/v2/views",
	"write":     "/api/v2/write",
}

func (h *APIHandler) serveLinks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := encodeResponse(ctx, w, http.StatusOK, apiLinks); err != nil {
		EncodeError(ctx, err, w)
		return
	}
}

// ServeHTTP delegates a request to the appropriate subhandler.
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSResponseHeaders(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	// Serve the links base links for the API.
	if r.URL.Path == "/api/v2/" || r.URL.Path == "/api/v2" {
		h.serveLinks(w, r)
		return
	}

	if r.URL.Path == "/api/v2/signin" || r.URL.Path == "/api/v2/signout" {
		h.SessionHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/setup") {
		h.SetupHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/write") {
		h.WriteHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/query") {
		h.QueryHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/buckets") {
		h.BucketHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/users") {
		h.UserHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/me") {
		h.UserHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/orgs") {
		h.OrgHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/authorizations") {
		h.AuthorizationHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/dashboards") {
		h.DashboardHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/sources") {
		h.SourceHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/tasks") {
		h.TaskHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/telegrafs") {
		h.TelegrafHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/views") {
		h.ViewHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v2/macros") {
		h.MacroHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/chronograf/") {
		h.ChronografHandler.ServeHTTP(w, r)
		return
	}

	notFoundHandler(w, r)
}
