// +build appenginevm

package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/rutmir/go-core/common"
	"github.com/rutmir/go-core/logger"
	"google.golang.org/appengine"
)

func createMux() *echo.Echo {
	// initialize internal utils
	logger.Initialize(
		common.GetRequiredProperty("LOG_PATH_SEPARATOR"),
		common.GetRequiredProperty("LOG_TARGET"))

	e := echo.New()
	// note: we don't need to provide the middleware or static handlers
	// for the appengine vm version - that's taken care of by the platform
	return e
}

func main() {
	// the appengine package provides a convenient method to handle the health-check requests
	// and also run the app on the correct port. We just need to add Echo to the default handler
	e := echo.New(":8080")
	http.Handle("/", e)
	appengine.Main()
}
