package app

import (
	"context"
	"net/http"
	"time"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/handler"
	"github.com/NailUsmanov/linkchecker11_11_2025/internal/storage"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// App - состоит из маршуртизатора chi, храншлища, логгера.
type App struct {
	router  *chi.Mux
	storage storage.Storage
	sugar   *zap.SugaredLogger
}

// NewApp - создадим новую стркутуру Арр.
// В ней регистрируем маршруты.
func NewApp(s storage.Storage, sugar *zap.SugaredLogger) *App {
	r := chi.NewRouter()
	app := &App{
		router:  r,
		storage: s,
		sugar:   sugar,
	}
	app.setupRoutes()
	return app
}

func (a *App) setupRoutes() {
	a.router.Post("/links", handler.NewCreateLinks(a.storage, a.sugar))
	a.router.Get("/links_num", handler.NewGetLinks(a.storage, a.sugar))
}

// Run будет запускать HTTP-сервер на указаноом адресе
func (a *App) Run(ctx context.Context, addr string) error {
	srv := http.Server{
		Addr:    addr,
		Handler: a.router,
	}

	go func() {
		<-ctx.Done()
		a.sugar.Infof("Shutdown the server")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
