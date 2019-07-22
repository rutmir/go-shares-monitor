package crawler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo"
	"github.com/rutmir/go-core/logger"
)

// Endpoint ...
type Endpoint struct {
	apiVersion    string
	sourceBaseURL string
}

// NewEndpoint ...
func NewEndpoint(apiVersion, sourceBaseURL, envName string, dbConfig *DBConfig) (*Endpoint, error) {
	if len(apiVersion) == 0 {
		err := errors.New("'apiVersion' param required")
		logger.ErrTrace(err)
		return nil, err
	}

	if len(sourceBaseURL) == 0 {
		err := errors.New("'sourceBaseURL' param required")
		logger.ErrTrace(err)
		return nil, err
	}

	environment = envName

	err := dbInit(dbConfig)
	if err != nil {
		return nil, err
	}

	return &Endpoint{apiVersion, sourceBaseURL}, nil
}

// RefreshIssuerlistHandler ...
func (h *Endpoint) RefreshIssuerlistHandler(c echo.Context) error {
	list, err := fetchIssuerlist(h.sourceBaseURL)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	issuers := make(map[string]string)
	for _, item := range list {
		if len(item.SLKey) > 0 {
			issuers[item.SLKey] = item.SLName
		}
	}

	err = updateIssuers(issuers)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	err = updateTickers(list)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
