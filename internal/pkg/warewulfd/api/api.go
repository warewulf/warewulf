package api

import (
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v5emb"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/version"
)

func Handler(auth *config.Authentication, allowedNets []net.IPNet) *web.Service {
	api := web.NewService(openapi3.NewReflector())

	api.OpenAPISchema().SetTitle("Warewulf v4 API")
	api.OpenAPISchema().SetDescription("This service provides an API to a Warewulf v4 server.")
	api.OpenAPISchema().SetVersion(version.GetVersion())

	api.Route("/api/nodes", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(auth, allowedNets))

			r.Method(http.MethodGet, "/", nethttp.NewHandler(getNodes()))
			r.Method(http.MethodGet, "/{id}", nethttp.NewHandler(getNodeByID()))
			r.Method(http.MethodGet, "/{id}/raw", nethttp.NewHandler(getRawNodeByID()))
			r.Method(http.MethodPut, "/{id}", nethttp.NewHandler(addNode()))
			r.Method(http.MethodDelete, "/{id}", nethttp.NewHandler(deleteNode()))
			r.Method(http.MethodPatch, "/{id}", nethttp.NewHandler(updateNode()))
			r.Method(http.MethodGet, "/{id}/fields", nethttp.NewHandler(getNodeFields()))
			r.Method(http.MethodPost, "/overlays/build", nethttp.NewHandler(buildAllOverlays()))
			r.Method(http.MethodPost, "/{id}/overlays/build", nethttp.NewHandler(buildOverlays()))
			r.Method(http.MethodGet, "/{id}/overlays", nethttp.NewHandler(getNodeOverlayInfo()))
		})
	})

	api.Route("/api/profiles", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(auth, allowedNets))

			r.Method(http.MethodGet, "/", nethttp.NewHandler(getProfiles()))
			r.Method(http.MethodGet, "/{id}", nethttp.NewHandler(getProfileByID()))
			r.Method(http.MethodPut, "/{id}", nethttp.NewHandler(addProfile()))
			r.Method(http.MethodPatch, "/{id}", nethttp.NewHandler(updateProfile()))
			r.Method(http.MethodDelete, "/{id}", nethttp.NewHandler(deleteProfile()))
		})
	})

	api.Route("/api/images", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(auth, allowedNets))

			r.Method(http.MethodGet, "/", nethttp.NewHandler(getImages()))
			r.Method(http.MethodGet, "/{name}", nethttp.NewHandler(getImageByName()))
			r.Method(http.MethodPost, "/{name}/import", nethttp.NewHandler(importImage()))
			r.Method(http.MethodPatch, "/{name}", nethttp.NewHandler(updateImage()))
			r.Method(http.MethodPost, "/{name}/build", nethttp.NewHandler(buildImage()))
			r.Method(http.MethodDelete, "/{name}", nethttp.NewHandler(deleteImage()))
		})
	})

	api.Route("/api/overlays", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(auth, allowedNets))

			r.Method(http.MethodGet, "/", nethttp.NewHandler(getOverlays()))
			r.Method(http.MethodGet, "/{name}", nethttp.NewHandler(getOverlayByName()))
			r.Method(http.MethodGet, "/{name}/file", nethttp.NewHandler(getOverlayFile()))
			r.Method(http.MethodPut, "/{name}", nethttp.NewHandler(createOverlay()))
			r.Method(http.MethodDelete, "/{name}", nethttp.NewHandler(deleteOverlay()))
		})
	})

	api.Docs("/api/docs", swgui.New)

	return api
}
