package scraping

import (
	"be/pkg/config"
	u "be/utils"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-shiori/go-readability"
	"github.com/pocketbase/pocketbase/daos"
)

func GetArticle(dao *daos.Dao, pageUrl string, onlyArticle bool) (*ParsedPage, bool, error) {

	// get html
	html, withProxy, err := getHtml(dao, pageUrl, true)
	if err != nil {
		return nil, withProxy, err
	}

	article, err := GetArticleFromHtml(html, pageUrl)
	if err != nil {
		return nil, withProxy, err
	}

	if onlyArticle {
		// get text in body
		textInBody, err := getTextInBody(html)
		if err != nil {
			return nil, withProxy, err
		}
		article.TextContent = textInBody
	}

	return article, withProxy, nil
}

// actually just builds a struct that represents the html
func GetArticleFromHtml(html string, pageUrl string) (*ParsedPage, error) {

	// create url struct
	pageUrlStruct, err := url.Parse(pageUrl)
	if err != nil {
		return nil, u.WrapError("cant parse url", err)
	}

	article, err := readability.FromReader(strings.NewReader(html), pageUrlStruct)
	if err != nil {
		return nil, u.WrapError("failed to parse html", err)
	}

	pp := ParsedPage{
		Title:       article.Title,
		TextContent: article.TextContent,
	}

	return &pp, nil
}

func getTextInBody(pageHtml string) (string, error) {

	h := html{}
	err := xml.NewDecoder(bytes.NewBufferString(pageHtml)).Decode(&h)
	if err != nil {
		return "", u.WrapError("failed to decode html", err)
	}

	return h.Body.Content, nil
}

func getProxyUrl(dao *daos.Dao) (*url.URL, error) {
	// set proxy
	proxy, err := config.GetRandomProxy(dao)
	if err != nil {
		return nil, err
	}
	// prepend http:// to proxy address
	// if not present
	if !strings.HasPrefix(proxy.Address, "http://") {
		proxy.Address = "http://" + proxy.Address
	}
	proxyUrl, err := url.Parse(string(proxy.Address) + ":" + fmt.Sprint(proxy.Port))
	if err != nil {
		return nil, err
	}
	return proxyUrl, nil
}

// function that takes a url, makes an http request to it, and returns the html
// if the request is done with proxy, it returns true as second return value
// pageUrl: url to make request to

func getHtml(dao *daos.Dao, pageUrl string, useProxy bool) (string, bool, error) {

	if useProxy {
		proxyUrl, err := getProxyUrl(dao)
		if err != nil {
			return "", useProxy, err
		}
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	} else {
		http.DefaultTransport = &http.Transport{Proxy: nil}
	}

	// make http request
	// prepend http:// or https://  to pageUrl if not present
	if !strings.HasPrefix(pageUrl, "http://") && !strings.HasPrefix(pageUrl, "https://") {
		pageUrl = "http://" + pageUrl
	}
	resp, err := http.Get(pageUrl)
	if err != nil {
		return "", useProxy, err
	}
	defer resp.Body.Close()

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", useProxy, err
	}

	// convert []byte to string
	html := string(body)
	return html, useProxy, nil
}
