package app

import (
	"github.com/go-ap/errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/openshift/osin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (f FedBOX) CollectionRoutes(descend bool) func (chi.Router) {
	return func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.With(middleware.GetHead)

			r.Method(http.MethodGet, "/", HandleCollection(f))
			r.Method(http.MethodPost, "/", HandleRequest(f))

			r.Route("/{id}", func(r chi.Router) {
				r.Method(http.MethodGet, "/", HandleItem(f))
				if descend {
					r.Route("/{collection}", f.CollectionRoutes(false))
				}
			})
		})
	}
}

func (f FedBOX) Routes(baseURL string, os *osin.Server, l logrus.FieldLogger) func(chi.Router) {
	return func(r chi.Router) {
		r.Use(middleware.RealIP)
		r.Use(middleware.GetHead)
		r.Use(ActorFromAuthHeader(os, f.Storage, l))

		r.Method(http.MethodGet, "/", HandleItem(f))
		r.Route("/{collection}", f.CollectionRoutes(true))

		h := oauthHandler{
			baseURL: baseURL,
			os:      os,
			loader:  f.Storage,
			logger:  l,
		}
		r.Route("/oauth", func(r chi.Router) {
			// Authorization code endpoint
			r.Get("/authorize", h.Authorize)
			r.Post("/authorize", h.Authorize)
			// Access token endpoint
			r.Post("/token", h.Token)

			r.Group(func(r chi.Router) {
				r.Get("/login", h.ShowLogin)
				r.Post("/login", h.HandleLogin)
				r.Get("/callback", h.HandleCallback)
				r.Get("/pw", h.ShowChangePw)
				r.Post("/pw", h.HandleChangePw)
			})
		})

		r.NotFound(errors.HandleError(errors.NotFoundf("invalid url")).ServeHTTP)
		r.MethodNotAllowed(errors.HandleError(errors.MethodNotAllowedf("method not allowed")).ServeHTTP)
	}
}
