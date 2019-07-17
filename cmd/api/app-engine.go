// +build appengine

package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/rutmir/go-core/common"
	"github.com/rutmir/go-core/logger"
)

func createMux() *echo.Echo {
	// initialize internal utils
	logger.Initialize(
		common.GetRequiredProperty("LOG_PATH_SEPARATOR"),
		common.GetRequiredProperty("LOG_TARGET"))

	e := echo.New()
	// note: we don't need to provide the middleware or static handlers, that's taken care of by the platform
	// app engine has it's own "main" wrapper - we just need to hook echo into the default handler
	http.Handle("/", e)
	return e
}
