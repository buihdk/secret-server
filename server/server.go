package server

import (
	"secretserver/secret"

	"github.com/labstack/echo"
)

func Serve() {
	s := echo.New()

	// config api route
	serve(s)

	s.Logger.Fatal(s.Start(":8080"))
}

func serve(s *echo.Echo) {
	s.POST("/secret", secret.AddSecret)
	s.GET("/secret/:hash", secret.GetSecret)
}
