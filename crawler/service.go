package crawler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo"
	"github.com/rutmir/go-core/logger"
)

type Endpoint struct {
	apiVersion string
}

func NewEndpoint(apiVersion string) (*Endpoint, error) {
	if len(apiVersion) == 0 {
		err := errors.New("'apiVersion' param required")
		logger.ErrTrace(err)
		return nil, err
	}

	return &Endpoint{apiVersion: apiVersion}, nil
}

func (h *Endpoint) RefreshIssuerlistHandler(c echo.Context) error {
	_ = fetchIssuerlist()
	return c.NoContent(http.StatusOK)
}
