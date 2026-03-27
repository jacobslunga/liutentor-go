package main

import (
	"log"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"liutentor-go-api/internal/config"
	"liutentor-go-api/internal/db"
	examhandler "liutentor-go-api/internal/handler/exam"
)

func main() {
	cfg := config.Load()
	supabase, err := db.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseServiceKey)
	if err != nil {
		log.Fatal("Failed to init Supabase ", err)
	}

	examH := examhandler.NewHandler(supabase)

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Only allow LiU Tentor hosts
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://liutentor.se", "http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "DELETE"},
	}))

	v1 := e.Group("/v1")

	exams := v1.Group("/exams")
	exams.GET("/:university/:courseCode", examH.GetExams)
	exams.GET("/:examId", examH.GetExam)

	port := cfg.Port
	if port == "" {
		port = "1323"
	}

	e.Start(":" + port)
}
