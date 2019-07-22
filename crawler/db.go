package crawler

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"runtime"
	"strings"
	"time"

	"github.com/go-pg/pg"
	"github.com/google/uuid"
	"github.com/rutmir/go-core/logger"
)

// DBConfig ...
type DBConfig struct {
	URL            string `json:"url"`
	User           string `json:"user"`
	Password       string `json:"password"`
	DBName         string `json:"dbname"`
	SSLMode        bool   `json:"sslmode"`
	PoolSize       int    `json:"poolSize"`
	CaCertFilePath string `json:"caCertFilePath"`
	CertFilePath   string `json:"certFilePath"`
	KeyFilePath    string `json:"keyFilePath"`
}

const (
	queryStartTime = "StartTime"
	dbServiceName  = "db"
)

var (
	db       *pg.DB
	cfg      *DBConfig
	certPool *x509.CertPool
	certs    []tls.Certificate
)

type queryLogger struct {
}

func (p queryLogger) BeforeQuery(context context.Context, event *pg.QueryEvent) (context.Context, error) {
	if event.Stash == nil {
		event.Stash = make(map[interface{}]interface{})
	}
	event.Stash[queryStartTime] = time.Now()

	return context, nil
}

func (p queryLogger) AfterQuery(context context.Context, event *pg.QueryEvent) (context.Context, error) {
	query, err := event.FormattedQuery()
	if err != nil {
		logger.Fatal(err)
	}

	if event.Stash != nil {
		if v, ok := event.Stash[queryStartTime]; ok {
			logger.Debugf("%s %s", time.Since(v.(time.Time)).Truncate(time.Second), query)
			return context, nil
		}
	}

	logger.Debug(query)
	return context, nil
}

// Connect connects to the DB
func Connect() {
	if db != nil {
		return
	}

	if cfg.PoolSize < 10 {
		numCPU := runtime.NumCPU()
		if numCPU < 1 {
			numCPU = 1
		}

		cfg.PoolSize = numCPU * 20
	}

	logger.Infof("connecting to %s, db %s, user %s, ssl %v, poolSize %v", cfg.URL, cfg.DBName, cfg.User, cfg.SSLMode, cfg.PoolSize)
	options := &pg.Options{
		Addr:     cfg.URL,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.DBName,
		PoolSize: cfg.PoolSize,
	}

	if cfg.SSLMode {
		options.TLSConfig = &tls.Config{
			Certificates: certs,
			RootCAs:      certPool,
			ServerName:   dbServiceName,
		}
	}

	db = pg.Connect(options)
	if strings.EqualFold(devEnvironmentName, environment) {
		hook := queryLogger{}
		db.AddQueryHook(hook)
	}
}

func dbInit(config *DBConfig) error {
	if config == nil {
		err := errors.New("'DBConfig' required")
		logger.Err(err)
		return err
	}

	cfg = config

	if !config.SSLMode {
		return nil
	}

	tlsCrt, err := ioutil.ReadFile(config.CertFilePath)
	if err != nil {
		logger.Err(err)
		return err
	}

	tlsKey, err := ioutil.ReadFile(config.KeyFilePath)
	if err != nil {
		logger.Err(err)
		return err
	}

	cert, err := tls.X509KeyPair(tlsCrt, tlsKey)
	if err != nil {
		logger.Err(err)
		return err
	}

	certs = []tls.Certificate{cert}

	caCrt, err := ioutil.ReadFile(config.CaCertFilePath)
	if err != nil {
		logger.Err(err)
		return err
	}
	certPool = x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCrt)

	return nil
}

// Issuer ...
type Issuer struct {
	ID       uuid.UUID `sql:",type:uuid"`
	Name     string
	SLKey    string `sql:",notnull"`
	SectorID *int
}

// Ticker ...
type Ticker struct {
	TickerSymbol string    `sql:",notnull"`
	IssuerID     uuid.UUID `sql:",notnull"`
	SLName       string    `sql:",notnull"`
}

func updateIssuers(keys map[string]string) error {
	Connect()

	issuers := []Issuer{}
	for k, v := range keys {
		issuers = append(issuers, Issuer{
			SLKey: k,
			Name:  v,
		})
	}

	_, err := db.Model(&issuers).OnConflict("DO NOTHING").Insert()
	if err != nil {
		logger.ErrTrace(err)
	}

	return err
}

func updateTickers(source []issuer) error {
	Connect()

	var issuers []Issuer
	err := db.Model(&issuers).Select()
	if err != nil {
		logger.ErrTrace(err)
		return err
	}

	issuersMap := make(map[string]uuid.UUID)
	for _, item := range issuers {
		issuersMap[item.SLKey] = item.ID
	}

	tickers := []Ticker{}
	for _, item := range source {
		if issuerID, ok := issuersMap[item.SLKey]; ok {
			tickers = append(tickers, Ticker{
				IssuerID:     issuerID,
				TickerSymbol: item.TickerSymbol,
				SLName:       item.SLName,
			})
		}
	}

	_, err = db.Model(&tickers).OnConflict("DO NOTHING").Insert()
	if err != nil {
		logger.ErrTrace(err)
	}

	return err
}
