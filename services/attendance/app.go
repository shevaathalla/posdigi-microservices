package main

import (
	"posdigi-attendance/config"
	"posdigi-attendance/database"
	"posdigi-attendance/handler"
	"posdigi-attendance/model"
	"posdigi-attendance/repository"
	"posdigi-attendance/router"
	"posdigi-attendance/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// App represents the application structure
type App struct {
	Config            *config.Config
	Router            *echo.Echo
	Logger            *logrus.Logger
	AttendanceHandler *handler.AttendanceHandler
}

// Bootstrap initializes the application
func Bootstrap() (*App, error) {
	// Initialize logger
	log := config.InitLogger()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.InitPostgres(cfg)
	if err != nil {
		return nil, err
	}

	// Auto-migrate database schema
	if err := db.AutoMigrate(&model.Attendance{}); err != nil {
		return nil, err
	}

	// Initialize layers
	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceService := service.NewAttendanceService(attendanceRepo, cfg)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)

	// Setup router
	e := router.Setup(log, attendanceHandler)

	return &App{
		Config:            cfg,
		Router:            e,
		Logger:            log,
		AttendanceHandler: attendanceHandler,
	}, nil
}
