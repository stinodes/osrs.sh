package wiki

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"osrs.sh/wiki/ssh/src/cmd"
)

type QueryResult struct {
	Query struct {
		Search []struct {
			Title   string `json:"title"`
			PageID  int    `json:"pageid"`
			Snippet string `json:"snippet"`
		} `json:"search"`
	} `json:"query"`
}

type ParseResult struct {
	Parse Page `json:"parse"`
}
type Page struct {
	Title      string `json:"title"`
	PageID     int    `json:"pageid"`
	Categories []struct {
		Category string `json:"category"`
	} `json:"categories"`
	Sections []struct {
		TocLevel int    `json:"toclevel"`
		Level    string `json:"level"`
		Line     string `json:"line"`
		Index    string `json:"index"`
	} `json:"sections"`
	WikiText string `json:"wikitext"`
}

func getHttpClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10,
	}
}

func searchUrl(query string) string {
	baseUrl := "https://oldschool.runescape.wiki/api.php?action=query&format=json&list=search&redirects=1&formatversion=2&srprop=size%7Cwordcount%7Ctimestamp%7Csnippet"
	searchParam := "srsearch=" + query
	return baseUrl + "&" + searchParam
}
func Search(query string) (*QueryResult, error) {
	log.Info("wiki", "query", query)

	client := getHttpClient()

	res, err := client.Get(searchUrl(query))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	result := QueryResult{}
	return &result, json.NewDecoder(res.Body).Decode(&result)
}

// https://oldschool.runescape.wiki/api.php?action=parse&format=json&pageid=44134&prop=categories%7Csections%7Crevid%7Cdisplaytitle%7Ciwlinks%7Cproperties%7Cparsewarnings%7Cwikitext&formatversion=2
func pageUrl(msg cmd.OpenArticle) string {
	baseUrl := "https://oldschool.runescape.wiki/api.php?action=parse&format=json&prop=categories%7Csections%7Crevid%7Cdisplaytitle%7Ciwlinks%7Cproperties%7Cparsewarnings%7Cwikitext&formatversion=2"
	var searchParam string
	if msg.PageId != 0 {
		searchParam = "pageid=" + strconv.Itoa(msg.PageId)
	} else {
		searchParam = "page=" + msg.Name
	}
	return baseUrl + "&" + searchParam
}
func ParsePage(msg cmd.OpenArticle) (*Page, error) {
	client := getHttpClient()

	res, err := client.Get(pageUrl(msg))
	if err != nil {
		log.Error("wiki", "err", err)
		return nil, err
	}

	defer res.Body.Close()

	result := ParseResult{}
	err = json.NewDecoder(res.Body).Decode(&result)

	return &result.Parse, err
}
