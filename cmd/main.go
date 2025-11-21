package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/app"
	"github.com/NailUsmanov/linkchecker11_11_2025/internal/storage"
	"go.uber.org/zap"
)

func main() {
	// Запускаю логирование
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	// Запуск регистратора
	sugar := logger.Sugar()

	// Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// хранилище для ссылок
	store, err := storage.NewFileStorage("data.json")
	if err != nil {
		sugar.Fatalf("create file storage failed: %v", err)
	}

	// создаем арр
	applictaion := app.NewApp(store, sugar)

	// Логирую запуск сервера и вызывваю Run
	sugar.Infow("starting HTTP server", "addr", ":8080")
	if err := applictaion.Run(ctx, ":8080"); err != nil {
		sugar.Fatalln(err)
	}
	sugar.Infow("server stop")

}
