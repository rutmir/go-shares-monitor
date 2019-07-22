package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rutmir/go-core/logger"
	"golang.org/x/net/html"
)

const (
	table = "table"
	tr    = "tr"
	th    = "th"
	td    = "td"
	a     = "a"

	name   = "название"
	ticker = "тикер"

	devEnvironmentName = "dev"
)

type issuer struct {
	SLName       string
	SLKey        string
	TickerSymbol string
}

var (
	environment string
)

func processErrorToken(err error) error {
	if err == io.EOF {
		return nil
	}
	logger.ErrTrace(err)

	return err
}

func detectColumnIndex(columnMap map[string]int, text string, index int) {
	switch text {
	case name:
		columnMap[name] = index
	case ticker:
		columnMap[ticker] = index
	}
}

func getHref(t html.Token) (bool, string) {
	for _, a := range t.Attr {
		if strings.EqualFold("href", a.Key) {
			return true, a.Val
		}
	}
	return false, ""
}

func getSLKey(key string) (string, error) {
	decoded, err := url.QueryUnescape(key)
	if err != nil {
		logger.ErrTrace(err)
		return "", err
	}

	idx := strings.LastIndex(decoded, "/")
	if idx >= 0 {
		return decoded[idx+1:], nil
	}

	return decoded, nil
}

func parseIssuerTableTitle(z *html.Tokenizer) (map[string]int, error) {
	index := -1
	result := make(map[string]int)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return nil, processErrorToken(z.Err())
		case html.TextToken:
			if index < 0 {
				continue
			}
			detectColumnIndex(result, strings.TrimSpace(strings.ToLower(string(z.Text()))), index)
		case html.StartTagToken:
			tn, _ := z.TagName()
			if strings.EqualFold(string(tn), th) {
				index++
			}
		case html.EndTagToken:
			tn, _ := z.TagName()
			if strings.EqualFold(string(tn), tr) {
				return result, nil
			}
		}
	}
}

func parseIssuerName(z *html.Tokenizer) (string, string, error) {
	slName := ""
	slKey := ""
	var err error

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return slName, slKey, processErrorToken(z.Err())
		case html.TextToken:
			slName = strings.TrimSpace(string(z.Text()))
		case html.StartTagToken:
			token := z.Token()
			if strings.EqualFold(token.Data, a) {
				ok, key := getHref(token)
				if ok {
					slKey, err = getSLKey(key)
					if err != nil {
						return slName, slKey, err
					}
				}
			}
		case html.EndTagToken:
			tn, _ := z.TagName()
			if strings.EqualFold(string(tn), td) {
				return slName, slKey, nil
			}
		}
	}
}

func parseIssuerTicker(z *html.Tokenizer) (string, error) {
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return "", processErrorToken(z.Err())
		case html.TextToken:
			return strings.TrimSpace(string(z.Text())), nil
		}
	}
}

func parseIssuerRowData(z *html.Tokenizer, indexes map[string]int, currentIndex int) (issuer, error) {
	result := issuer{}
	var err error

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return result, processErrorToken(z.Err())
		case html.StartTagToken:
			tn, _ := z.TagName()
			switch strings.ToLower(string(tn)) {
			case td:
				currentIndex++
				for k, v := range indexes {
					if v != currentIndex {
						continue
					}

					switch k {
					case name:
						result.SLName, result.SLKey, err = parseIssuerName(z)
						if err != nil {
							return result, err
						}

					case ticker:
						result.TickerSymbol, err = parseIssuerTicker(z)
						if err != nil {
							return result, err
						}
					}
				}
			}
		case html.EndTagToken:
			tn, _ := z.TagName()
			if strings.EqualFold(string(tn), tr) {
				return result, nil
			}
		}
	}
}

func parseIssuerTableData(z *html.Tokenizer, indexes map[string]int) ([]issuer, error) {
	result := []issuer{}
	needText := false
	index := -1
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return nil, processErrorToken(z.Err())
		case html.TextToken:
			if !needText {
				continue
			}
			needText = false
			if len(strings.TrimSpace(string(z.Text()))) > 0 {
				issuer, err := parseIssuerRowData(z, indexes, index)
				if err != nil {
					return nil, err
				}

				result = append(result, issuer)
			}
		case html.StartTagToken:
			tn, _ := z.TagName()
			switch strings.ToLower(string(tn)) {
			case tr:
				index = -1
			case td:
				index++
				if index != 0 {
					needText = false
					continue
				}
				needText = true
			}
		case html.EndTagToken:
			tn, _ := z.TagName()
			if strings.EqualFold(string(tn), table) {
				return result, nil
			}
		}
	}
}

func parseIssuerList(z *html.Tokenizer) ([]issuer, error) {
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return nil, processErrorToken(z.Err())
		case html.StartTagToken:
			tn, _ := z.TagName()
			if strings.EqualFold(string(tn), table) {
				idx, err := parseIssuerTableTitle(z)
				if err != nil {
					return nil, err
				}
				fmt.Println(idx)
				return parseIssuerTableData(z, idx)
			}
		}
	}
}

func fetchIssuerlist(baseURL string) ([]issuer, error) {
	url := fmt.Sprintf("%s/q/shares/", baseURL)

	logger.Infof("HTML code of %s ...", url)
	resp, err := http.Get(url)
	if err != nil {
		logger.ErrTrace(err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("request to '%s' failed", url)
		logger.ErrTrace(err)
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	list, err := parseIssuerList(html.NewTokenizer(resp.Body))
	if err != nil {
		return nil, err
	}

	return list, nil
}
