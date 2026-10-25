package server

import (
	"net/http"
	"os"

	"secretserver/internal/crypto"
	"secretserver/internal/metrics"
	"secretserver/secret"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Serve() {
	s := echo.New()

	// Refuse to start without a real encryption key. There is no default.
	if err := crypto.CheckKey(); err != nil {
		s.Logger.Fatal(err)
	}

	s.Use(middleware.Logger())
	s.Use(middleware.Recover())
	s.Use(metrics.Middleware())

	serve(s)

	s.Logger.Fatal(s.Start(":8080"))
}

func serve(s *echo.Echo) {
	s.GET("/metrics", metricsHandler())
	s.POST("/secret", secret.AddSecret)
	s.GET("/secret/:hash", secret.GetSecret)
}

// metricsHandler wraps the metrics handler with a simple token guard.
// Set METRICS_TOKEN to require a matching Authorization Bearer token.
// If METRICS_TOKEN is unset the endpoint stays open for backward compatibility.
func metricsHandler() echo.HandlerFunc {
	h := metrics.Handler()
	token := os.Getenv("METRICS_TOKEN")
	if token == "" {
		return h
	}
	return func(c echo.Context) error {
		if c.Request().Header.Get("Authorization") != "Bearer "+token {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}
		return h(c)
	}
}
