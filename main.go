package main

import (
	"log"
	"net/http"

	"github.com/madza/tubley/config"
	"github.com/madza/tubley/handlers"
	"github.com/madza/tubley/middleware"

	_ "github.com/lib/pq"
)

func main() {

	cfg := config.NewApiConfig()

	err := cfg.EnsureAssetsDir()
	if err != nil {
		log.Fatalf("Couldn't create assets directory: %v", err)
	}

	mux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(cfg.FilepathRoot)))
	mux.Handle("/app/", appHandler)

	assetsHandler := http.StripPrefix("/assets", http.FileServer(http.Dir(cfg.AssetsRoot)))
	mux.Handle("/assets/", middleware.NoCacheMiddleware(assetsHandler))

	mux.HandleFunc("POST /api/login", handlers.HandlerLogin(cfg))
	mux.HandleFunc("POST /api/refresh", handlers.HandlerRefresh((cfg)))
	mux.HandleFunc("POST /api/revoke", handlers.HandlerRevoke((cfg)))

	mux.HandleFunc("POST /api/users", handlers.HandlerUsersCreate(cfg))

	mux.HandleFunc("POST /api/videos", handlers.HandlerVideoMetaCreate(cfg))
	mux.HandleFunc("POST /api/thumbnail_upload/{videoID}", handlers.HandlerUploadThumbnail(cfg))
	mux.HandleFunc("POST /api/video_upload/{videoID}", handlers.HandlerUploadVideo(cfg))
	mux.HandleFunc("GET /api/videos", handlers.HandlerVideosRetrieve(cfg))
	mux.HandleFunc("GET /api/videos/{videoID}", handlers.HandlerVideoGet(cfg))
	//mux.HandleFunc("GET /api/thumbnails/{videoID}", cfg.HandlerThumbnailGet)
	mux.HandleFunc("DELETE /api/videos/{videoID}", handlers.HandlerVideoMetaDelete(cfg))

	mux.HandleFunc("POST /admin/reset", handlers.HandlerReset(cfg))

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	log.Printf("Serving on: http://localhost:%s/app/\n", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
