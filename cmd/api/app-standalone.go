// +build !appengine,!appenginevm

package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/rutmir/go-core/common"
	"github.com/rutmir/go-core/logger"
	"github.com/rutmir/go-core/secret"
)

const (
	healthz     = "/healthz"
	defaultPort = "10443"
)

func createMux() *echo.Echo {
	// initialize internal utils
	logger.Initialize(
		common.GetRequiredProperty("LOG_PATH_SEPARATOR"),
		common.GetRequiredProperty("LOG_TARGET"))
	secret.Initialize(common.GetRequiredProperty("SECRETS_PATH"))

	// Init Echo
	e := echo.New()
	// e.Pre(middleware.HTTPSRedirect())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
		Skipper: func(c echo.Context) bool {
			return strings.EqualFold(c.Request().URL.Path, healthz)
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.GET(healthz, func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return e
}

func main() {
	port := common.GetPropertyOrDefault("SERVER_PORT", defaultPort)
	// caCertPath := secret.GetSecretFilePathOrPanic(common.GetRequiredProperty("CA_CERT_FILENAME"))
	certPath := secret.GetSecretFilePathOrPanic(common.GetRequiredProperty("CERT_FILENAME"))
	keyPath := secret.GetSecretFilePathOrPanic(common.GetRequiredProperty("KEY_FILENAME"))

	logger.Info("start HTTPS server on port: " + port)
	logger.Fatal(e.StartTLS(":"+port, certPath, keyPath))
}
