package main

import (
	"context"
	"github.com/Olzerq/avito-pr-reviewer/internal/config"

	"github.com/Olzerq/avito-pr-reviewer/internal/repository"
	"github.com/Olzerq/avito-pr-reviewer/internal/service"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"

	httpapi "github.com/Olzerq/avito-pr-reviewer/internal/http"
)

func main() {
	// Загрузка конфига
	cfg := config.LoadConfig()

	// Контекст для БД
	ctx := context.Background()

	// log.Printf("Connecting to DB: host=%s port=%s user=%s db=%s",
	//	cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName,
	// )

	// Подключаемся к БД
	db := repository.NewDB(ctx, cfg.DBConnStr())
	defer db.Close()

	teamRepo := repository.NewTeamRepository(db)
	teamService := service.NewTeamService(teamRepo)
	teamHandler := httpapi.NewTeamHandler(teamService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := httpapi.NewUserHandler(userService)

	prRepo := repository.NewPullRequestRepository(db)
	prService := service.NewPullRequestService(prRepo, userRepo)
	prHandler := httpapi.NewPullRequestHandler(prService)

	statsRepo := repository.NewStatsRepository(db)
	statsService := service.NewStatsService(statsRepo)
	statsHandler := httpapi.NewStatsHandler(statsService)

	// Создаем роутер
	r := chi.NewRouter()

	// Тестовый эндпоинт
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Добавление команды
	r.Post("/team/add", teamHandler.TeamAdd)

	// Получение команды
	r.Get("/team/get", teamHandler.TeamGet)

	// Обновление флага активности
	r.Post("/users/setIsActive", userHandler.SetIsActive)

	// Получает все PR'ы где юзер ревьювер
	r.Get("/users/getReview", prHandler.GetUserReviews)

	// Создание PR
	r.Post("/pullRequest/create", prHandler.Create)

	// Переназначение ревьюера
	r.Post("/pullRequest/reassign", prHandler.Reassign)

	// Флаг Merged
	r.Post("/pullRequest/merge", prHandler.Merge)

	// Статистика
	r.Get("/stats/assignments", statsHandler.GetAssignments)

	addr := ":" + cfg.AppPort
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
