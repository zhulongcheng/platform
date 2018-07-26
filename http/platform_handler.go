package http

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strings"

	idpctx "github.com/influxdata/platform/context"
	"github.com/prometheus/client_golang/prometheus"
)

// PlatformHandler is a collection of all the service handlers.
type PlatformHandler struct {
	BucketHandler        *BucketHandler
	UserHandler          *UserHandler
	OrgHandler           *OrgHandler
	AuthorizationHandler *AuthorizationHandler
	DashboardHandler     *DashboardHandler
	AssetHandler         *AssetHandler
	ChronografHandler    *ChronografHandler
	SourceHandler        *SourceHandler
	FluxLangHandler      *FluxLangHandler
}

func setCORSResponseHeaders(w nethttp.ResponseWriter, r *nethttp.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
	}
}

var platformRoutesResponse = mustMarshalJSON(map[string]interface{}{
	"sources": "/chronogaf/v1/sources",
	"flux": map[string]string{
		"self":        "/v2/flux",
		"ast":         "/v2/flux/ast",
		"suggestions": "/v2/flux/suggestions",
	},
	"chonograf":     "/chronograf/v1",
	"layouts":       "/chronograf/v1/layouts",
	"allUsers":      "/chronograf/v1/users",
	"organizations": "/chronograf/v1/organizations",
	"me":            "/chronograf/v1/me",
	"environment":   "/chronograf/v1/env",
	"mappings":      "/chronograf/v1/mappings",
	"dashboards":    "/chronograf/v1/dashboards",
	"dashboardsV2":  "/chronograf/v2/dashboards",
	"cells":         "/chronograf/v2/cells",
	"config": map[string]string{
		"self": "/chronograf/v1/config",
		"auth": "/chronograf/v1/config/auth",
	},
	"externalLinks": map[string]string{
		"statusFeed": "https://www.influxdata.com/feed/json",
	},
})

func (h *PlatformHandler) servePlatformRoutes(w nethttp.ResponseWriter, r *nethttp.Request) {
	w.WriteHeader(nethttp.StatusOK)
	w.Write(platformRoutesResponse)
}

// ServeHTTP delegates a request to the appropriate subhandler.
func (h *PlatformHandler) ServeHTTP(w nethttp.ResponseWriter, r *nethttp.Request) {

	setCORSResponseHeaders(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	// Server the chronograf assets for any basepath that does not start with addressable parts
	// of the platform API.
	if !strings.HasPrefix(r.URL.Path, "/v1/") &&
		!strings.HasPrefix(r.URL.Path, "/v2/") &&
		!strings.HasPrefix(r.URL.Path, "/chronograf/") {
		h.AssetHandler.ServeHTTP(w, r)
		return
	}

	// Serve the links base links for the API.
	if r.URL.Path == "/v2/" || r.URL.Path == "/v2" {
		h.servePlatformRoutes(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v2/flux") {
		h.FluxLangHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/chronograf/") {
		h.ChronografHandler.ServeHTTP(w, r)
		return
	}

	ctx := r.Context()
	var err error
	if ctx, err = extractAuthorization(ctx, r); err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusBadRequest)
		return
	}
	r = r.WithContext(ctx)

	if strings.HasPrefix(r.URL.Path, "/v1/buckets") {
		h.BucketHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v1/users") {
		h.UserHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v1/orgs") {
		h.OrgHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v1/authorizations") {
		h.AuthorizationHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v1/dashboards") {
		h.DashboardHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v2/sources") {
		h.SourceHandler.ServeHTTP(w, r)
		return
	}

	nethttp.NotFound(w, r)
}

// PrometheusCollectors satisfies the prom.PrometheusCollector interface.
func (h *PlatformHandler) PrometheusCollectors() []prometheus.Collector {
	// TODO: collect and return relevant metrics.
	return nil
}

func extractAuthorization(ctx context.Context, r *nethttp.Request) (context.Context, error) {
	t, err := ParseAuthHeaderToken(r)
	if err != nil {
		return ctx, err
	}
	return idpctx.SetToken(ctx, t), nil
}

func mustMarshalJSON(i interface{}) []byte {
	b, err := json.Marshal(i)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal json: %v", err))
	}

	return b
}
