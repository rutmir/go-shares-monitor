package main

import (
	"github.com/rutmir/go-core/logger"
	"github.com/rutmir/go-shares-monitor/crawler"
)

func init() {
	crawlerEP, err := crawler.NewEndpoint(apiVersion)
	if err != nil {
		logger.Fatal(err)
	}

	g := e.Group("/job")
	// g.Use(middleware.CORS())
	g.GET("/update-issuer-list", crawlerEP.RefreshIssuerlistHandler)
}
