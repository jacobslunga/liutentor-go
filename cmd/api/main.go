package main

import (
	"log"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"liutentor-go/internal/config"
	"liutentor-go/internal/db"
	coursehandler "liutentor-go/internal/handler/course"
	examhandler "liutentor-go/internal/handler/exam"
)

func main() {
	cfg := config.Load()
	supabase, err := db.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseServiceKey)
	if err != nil {
		log.Fatal("Failed to init Supabase ", err)
	}

	examH := examhandler.NewHandler(supabase)
	courseH := coursehandler.NewHandler(supabase)

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Only allow LiU Tentor hosts
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://liutentor.se", "http://localhost:5173", "http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "DELETE"},
	}))

	v1 := e.Group("/v1")

	exams := v1.Group("/exams")
	exams.GET("/:university/:courseCode", examH.GetExams)
	exams.GET("/:examId", examH.GetExam)

	courses := v1.Group("/courses")
	courses.GET("/:university", courseH.GetCourses)

	port := cfg.Port
	if port == "" {
		port = "1323"
	}

	if err := e.Start(":" + port); err != nil {
		log.Fatal("Failed to start server")
	}
}
