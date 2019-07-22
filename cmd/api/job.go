package main

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/rutmir/go-core/common"
	"github.com/rutmir/go-core/logger"
	"github.com/rutmir/go-core/secret"
	"github.com/rutmir/go-shares-monitor/crawler"
)

func init() {

	poolsPerCPU, err := strconv.Atoi(common.GetRequiredProperty("DB_POOLS_PER_CPU"))
	if err != nil {
		logger.Fatal(err)
	}

	numCPU := runtime.NumCPU()
	if numCPU < 1 {
		numCPU = 1
	}

	dbConfig := &crawler.DBConfig{
		URL:            secret.GetValueOrPanic("/db/db_conn"),
		User:           secret.GetValueOrPanic("/db/db_user"),
		Password:       secret.GetValueOrPanic("/db/db_pass"),
		DBName:         secret.GetValueOrPanic("/db/db_name"),
		SSLMode:        strings.EqualFold(secret.GetValueOrPanic("/db/db_ssl"), "true"),
		PoolSize:       numCPU * poolsPerCPU,
		CaCertFilePath: secret.GetSecretFilePathOrPanic(common.GetRequiredProperty("CA_CERT_FILENAME")),
		CertFilePath:   secret.GetSecretFilePathOrPanic(common.GetRequiredProperty("CERT_FILENAME")),
		KeyFilePath:    secret.GetSecretFilePathOrPanic(common.GetRequiredProperty("KEY_FILENAME")),
	}

	crawlerEP, err := crawler.NewEndpoint(apiVersion, common.GetRequiredProperty("SOURCE_BASE_URL"), "", dbConfig)
	if err != nil {
		logger.Fatal(err)
	}

	g := e.Group("/job")
	// g.Use(middleware.CORS())
	g.GET("/update-issuer-list", crawlerEP.RefreshIssuerlistHandler)
}
