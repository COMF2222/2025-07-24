package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"2025-07-24/internal/api"
	"2025-07-24/internal/config"
	"2025-07-24/internal/repository"
	"2025-07-24/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("ошибка загрузки конфига: %v", err)
	}

	repo := repository.NewInMemoryRepository()
	svc := service.NewService(repo, cfg)
	handler := api.NewHandler(svc)

	r := chi.NewRouter()
	r.Post("/tasks", handler.CreateTask)
	r.Post("/tasks/{id}/links", handler.AddLink)
	r.Get("/tasks/{id}/status", handler.GetStatus)
	r.Get("/archives/{id}", handler.DownloadArchive)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Сервер запущен на порте :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Сервер неисправен: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка выключения: %v", err)
	}
	log.Println("Сервер остановлен")
}
