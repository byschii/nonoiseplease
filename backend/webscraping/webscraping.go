package webscraping

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	config "be/model/config"
	proxy "be/model/proxy"
	u "be/utils"

	readability "github.com/go-shiori/go-readability"
	"github.com/pocketbase/pocketbase/daos"
)

func GetArticle(pageUrl string, onlyArticle bool, dao *daos.Dao) (*ParsedPage, bool, error) {

	// get html
	html, withProxy, err := getHtml(pageUrl, dao)
	if err != nil {
		return nil, withProxy, err
	}

	article, err := GetArticleFromHtml(html, pageUrl)
	if err != nil {
		return nil, withProxy, err
	}

	if onlyArticle {
		// get text in body
		textInBody, err := GetTextInBody(html)
		if err != nil {
			return nil, withProxy, err
		}
		article.TextContent = textInBody
	}

	return article, withProxy, nil
}

func GetTextInBody(pageHtml string) (string, error) {

	h := html{}
	err := xml.NewDecoder(bytes.NewBufferString(pageHtml)).Decode(&h)
	if err != nil {
		return "", u.WrapError("failed to decode html", err)
	}

	return h.Body.Content, nil
}

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

// function that takes a url, makes an http request to it, and returns the html
// if the request is done with proxy, it returns true as second return value
func getHtml(pageUrl string, dao *daos.Dao) (string, bool, error) {

	userProxyProb := config.GetConfigUseProxyProbability(dao)

	useProxy := rand.Float32() < userProxyProb
	log.Printf("useProxy: %v,  userProxyProb: %v", useProxy, userProxyProb)

	if useProxy {
		// set proxy
		proxy, err := proxy.GetRandomProxy(dao)
		if err != nil {
			return "", useProxy, err
		}
		// prepend http:// to proxy address
		// if not present
		if !strings.HasPrefix(proxy.Address, "http://") {
			proxy.Address = "http://" + proxy.Address
		}
		proxyUrl, err := url.Parse(string(proxy.Address) + ":" + fmt.Sprint(proxy.Port))
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", useProxy, err
	}

	// convert []byte to string
	html := string(body)

	return html, useProxy, nil
}
