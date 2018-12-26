package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil"
	"github.com/influxdata/platform"
	pctx "github.com/influxdata/platform/context"
	"github.com/influxdata/platform/kit/errors"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// TelegrafHandler is the handler for the telegraf service
type TelegrafHandler struct {
	*httprouter.Router
	Logger *zap.Logger

	TelegrafService platform.TelegrafConfigStore
}

const (
	telegrafsPath             = "/api/v2/telegrafs"
	telegrafsIDPath           = "/api/v2/telegrafs/:id"
	telegrafsIDMembersPath    = "/api/v2/telegrafs/:id/members"
	telegrafsIDMembersIDPath  = "/api/v2/telegrafs/:id/members/:userID"
	telegrafsIDOwnersPath     = "/api/v2/telegrafs/:id/owners"
	telegrafsIDOwnersIDPath   = "/api/v2/telegrafs/:id/owners/:userID"
	telegrafsIDLabelsPath     = "/api/v2/telegrafs/:id/labels"
	telegrafsIDLabelsNamePath = "/api/v2/telegrafs/:id/labels/:name"
)

// NewTelegrafHandler returns a new instance of TelegrafHandler.
func NewTelegrafHandler(b *APIBackend) *TelegrafHandler {
	h := &TelegrafHandler{
		Router: NewRouter(),
		Logger: b.Logger.With(zap.String("handler", "telegraf")),

		TelegrafService: b.TelegrafService,
	}
	h.HandlerFunc("POST", telegrafsPath, h.handlePostTelegraf)
	h.HandlerFunc("GET", telegrafsPath, h.handleGetTelegrafs)
	h.HandlerFunc("GET", telegrafsIDPath, h.handleGetTelegraf)
	h.HandlerFunc("DELETE", telegrafsIDPath, h.handleDeleteTelegraf)
	h.HandlerFunc("PUT", telegrafsIDPath, h.handlePutTelegraf)

	h.HandlerFunc("POST", telegrafsIDMembersPath, newPostMemberHandler(b, platform.TelegrafResourceType, platform.Member))
	h.HandlerFunc("GET", telegrafsIDMembersPath, newGetMembersHandler(b, platform.TelegrafResourceType, platform.Member))
	h.HandlerFunc("DELETE", telegrafsIDMembersIDPath, newDeleteMemberHandler(b, platform.Member))

	h.HandlerFunc("POST", telegrafsIDOwnersPath, newPostMemberHandler(b, platform.TelegrafResourceType, platform.Owner))
	h.HandlerFunc("GET", telegrafsIDOwnersPath, newGetMembersHandler(b, platform.TelegrafResourceType, platform.Owner))
	h.HandlerFunc("DELETE", telegrafsIDOwnersIDPath, newDeleteMemberHandler(b, platform.Owner))

	h.HandlerFunc("GET", telegrafsIDLabelsPath, newGetLabelsHandler(b.LabelService))
	h.HandlerFunc("POST", telegrafsIDLabelsPath, newPostLabelHandler(b.LabelService))
	h.HandlerFunc("DELETE", telegrafsIDLabelsNamePath, newDeleteLabelHandler(b.LabelService))
	h.HandlerFunc("PATCH", telegrafsIDLabelsNamePath, newPatchLabelHandler(b.LabelService))

	return h
}

type link struct {
	Self string `json:"self"`
}

type telegrafResponse struct {
	*platform.TelegrafConfig
	Links link `json:"links"`
}

type telegrafResponses struct {
	TelegrafConfigs []telegrafResponse `json:"configurations"`
}

func newTelegrafResponse(tc *platform.TelegrafConfig) telegrafResponse {
	return telegrafResponse{
		TelegrafConfig: tc,
		Links: link{
			Self: fmt.Sprintf("/api/v2/telegrafs/%s", tc.ID.String()),
		},
	}
}

func newTelegrafResponses(tcs []*platform.TelegrafConfig) telegrafResponses {
	resp := telegrafResponses{
		TelegrafConfigs: make([]telegrafResponse, len(tcs)),
	}
	for i, c := range tcs {
		resp.TelegrafConfigs[i] = newTelegrafResponse(c)
	}
	return resp
}

func decodeGetTelegrafRequest(ctx context.Context, r *http.Request) (i platform.ID, err error) {
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		return i, errors.InvalidDataf("url missing id")
	}

	if err := i.DecodeFromString(id); err != nil {
		return i, err
	}
	return i, nil
}

func (h *TelegrafHandler) handleGetTelegrafs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter, err := decodeUserResourceMappingFilter(ctx, r)
	if err != nil {
		h.Logger.Debug("failed to decode request", zap.Error(err))
		EncodeError(ctx, err, w)
		return
	}
	tcs, _, err := h.TelegrafService.FindTelegrafConfigs(ctx, *filter)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}
	if err := encodeResponse(ctx, w, http.StatusOK, newTelegrafResponses(tcs)); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

func (h *TelegrafHandler) handleGetTelegraf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := decodeGetTelegrafRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}
	tc, err := h.TelegrafService.FindTelegrafConfigByID(ctx, id)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	offers := []string{"application/toml", "application/json", "application/octet-stream"}
	defaultOffer := "application/toml"
	mimeType := httputil.NegotiateContentType(r, offers, defaultOffer)
	switch mimeType {
	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.toml\"", strings.Replace(strings.TrimSpace(tc.Name), " ", "_", -1)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tc.TOML()))
	case "application/json":
		if err := encodeResponse(ctx, w, http.StatusOK, newTelegrafResponse(tc)); err != nil {
			logEncodingError(h.Logger, r, err)
			return
		}
	case "application/toml":
		w.Header().Set("Content-Type", "application/toml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tc.TOML()))
	}
}

func decodeUserResourceMappingFilter(ctx context.Context, r *http.Request) (*platform.UserResourceMappingFilter, error) {
	q := r.URL.Query()
	f := &platform.UserResourceMappingFilter{
		ResourceType: platform.TelegrafResourceType,
	}
	if idStr := q.Get("resourceId"); idStr != "" {
		id, err := platform.IDFromString(idStr)
		if err != nil {
			return nil, err
		}
		f.ResourceID = *id
	}

	if idStr := q.Get("userId"); idStr != "" {
		id, err := platform.IDFromString(idStr)
		if err != nil {
			return nil, err
		}
		f.UserID = *id
	}
	return f, nil
}

func decodePostTelegrafRequest(ctx context.Context, r *http.Request) (*platform.TelegrafConfig, error) {
	tc := new(platform.TelegrafConfig)
	err := json.NewDecoder(r.Body).Decode(tc)
	return tc, err
}

func decodePutTelegrafRequest(ctx context.Context, r *http.Request) (*platform.TelegrafConfig, error) {
	tc := new(platform.TelegrafConfig)
	if err := json.NewDecoder(r.Body).Decode(tc); err != nil {
		return nil, err
	}
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		return nil, errors.InvalidDataf("url missing id")
	}
	i := new(platform.ID)
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}
	tc.ID = *i
	return tc, nil
}

// handlePostTelegraf is the HTTP handler for the POST /api/v2/telegrafs route.
func (h *TelegrafHandler) handlePostTelegraf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tc, err := decodePostTelegrafRequest(ctx, r)
	if err != nil {
		h.Logger.Debug("failed to decode request", zap.Error(err))
		EncodeError(ctx, err, w)
		return
	}
	auth, err := pctx.GetAuthorizer(ctx)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := h.TelegrafService.CreateTelegrafConfig(ctx, tc, auth.GetUserID()); err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusCreated, newTelegrafResponse(tc)); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

// handlePutTelegraf is the HTTP handler for the POST /api/v2/telegrafs route.
func (h *TelegrafHandler) handlePutTelegraf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tc, err := decodePutTelegrafRequest(ctx, r)
	if err != nil {
		h.Logger.Debug("failed to decode request", zap.Error(err))
		EncodeError(ctx, err, w)
		return
	}
	auth, err := pctx.GetAuthorizer(ctx)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	tc, err = h.TelegrafService.UpdateTelegrafConfig(ctx, tc.ID, tc, auth.GetUserID())
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusOK, newTelegrafResponse(tc)); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

func (h *TelegrafHandler) handleDeleteTelegraf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	i, err := decodeGetTelegrafRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err = h.TelegrafService.DeleteTelegrafConfig(ctx, i); err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusNoContent, nil); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}
