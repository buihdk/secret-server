package server

import (
	"secretserver/internal/metrics"
	"secretserver/secret"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func Serve() {
	s := echo.New()

	s.Use(middleware.Logger())
	s.Use(middleware.Recover())
	s.Use(metrics.Middleware())

	serve(s)

	s.Logger.Fatal(s.Start(":8080"))
}

func serve(s *echo.Echo) {
	s.GET("/metrics", metrics.Handler())
	s.POST("/secret", secret.AddSecret)
	s.GET("/secret/:hash", secret.GetSecret)
}
