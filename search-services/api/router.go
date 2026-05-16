package main

import (
	"log/slog"
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	httpSwagger "github.com/swaggo/http-swagger"
	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/adapters/rest/middleware"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
)

type Router struct {
	log          *slog.Logger
	cfg          config.Config
	updateClient core.Updater
	wordsClient  core.Normalizer
	searchClient core.Searcher
	aaa          core.AAA
	profile      core.Profile
}

func NewRouter(
	log *slog.Logger,
	cfg config.Config,
	updateClient core.Updater,
	wordsClient core.Normalizer,
	searchClient core.Searcher,
	aaa core.AAA,
	profile core.Profile,
) http.Handler {
	r := &Router{
		log:          log,
		cfg:          cfg,
		updateClient: updateClient,
		wordsClient:  wordsClient,
		searchClient: searchClient,
		aaa:          aaa,
		profile:      profile,
	}
	return r.router()
}

func (r *Router) router() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/ping", rest.NewPingHandler(r.log, map[string]core.Pinger{
		"words":  r.wordsClient,
		"update": r.updateClient,
		"search": r.searchClient,
		"aaa":    r.aaa,
	}))

	mux.Handle("POST /api/login", rest.NewLoginHandler(r.log, r.aaa))
	mux.Handle("POST /api/auth/register", rest.NewRegisterHandler(r.log, r.aaa))
	
	mux.Handle("GET /api/me", middleware.Auth(rest.NewProfileHandler(r.log, r.profile), r.aaa))
	mux.Handle("POST /api/me/saved/{id}", middleware.Auth(rest.NewSaveComicsHandler(r.log, r.profile), r.aaa))
	mux.Handle("DELETE /api/me/saved/{id}", middleware.Auth(rest.NewUnsaveComicsHandler(r.log, r.profile), r.aaa))
	mux.Handle("GET /api/comics/{id}", middleware.Rate(rest.NewComicsHandler(r.log, r.searchClient), r.cfg.SearchRate))

	mux.Handle("POST /api/db/update", middleware.Auth(middleware.RequireRole(core.RoleAdmin)(rest.NewUpdateHandler(r.log, r.updateClient)), r.aaa))
	mux.Handle("GET /api/db/stats", middleware.Auth(middleware.RequireRole(core.RoleAdmin)(rest.NewUpdateStatsHandler(r.log, r.updateClient)), r.aaa))
	mux.Handle("GET /api/db/status", middleware.Auth(middleware.RequireRole(core.RoleAdmin)(rest.NewUpdateStatusHandler(r.log, r.updateClient)), r.aaa))
	mux.Handle("DELETE /api/db", middleware.Auth(middleware.RequireRole(core.RoleAdmin)(rest.NewDropHandler(r.log, r.updateClient)), r.aaa))

	mux.Handle("GET /api/search", middleware.Concurrency(rest.NewSearchHandler(r.log, r.searchClient), r.cfg.SearchConcurrency))
	mux.Handle("GET /api/isearch", middleware.Rate(rest.NewISearchHandler(r.log, r.searchClient), r.cfg.SearchRate))
	

	mux.Handle("GET /metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, true)
	}))
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:28080/swagger/doc.json"),
	))

	return middleware.Logging(middleware.WithMetrics(mux), r.log)
}
