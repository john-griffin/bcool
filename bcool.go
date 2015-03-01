package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"code.google.com/p/go.net/html"
)

type Item struct {
	Link        string `xml:"link"`
	Title       string `xml:"title"`
	Creator     string `xml:"creator"`
	Guid        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Title   string   `xml:"title"`
	Items   []Item   `xml:"item"`
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Description struct {
	Key   int
	Value string
}

func OriginalFeedBody(category string) []byte {
	categoryPath := ""
	if len(category) > 0 {
		categoryPath = "/category/" + category
	}
	url := fmt.Sprintf("http://www.bleedingcool.com%s/feed/", categoryPath)
	log.Printf("Fetching %s", url)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func FetchFullDescription(link string) string {
	res, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	doc, err := html.Parse(strings.NewReader(string(body)))
	content := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "section" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "entry-content cf" {
					var buf bytes.Buffer
					html.Render(&buf, n)
					content = buf.String()
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return content
}

func Feed(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	body := OriginalFeedBody(category)
	v := Rss{}
	err := xml.Unmarshal(body, &v)
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan Description)
	for key, value := range v.Channel.Items {
		go func(key int, link string) {
			c <- Description{key, FetchFullDescription(link)}
		}(key, value.Link)
	}
	for _ = range v.Channel.Items {
		result := <-c
		v.Channel.Items[result.Key].Description = result.Value
	}
	b, err := xml.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, string(b))
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("/feed", Feed)
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	log.Printf("Starting on port %s ....", port)
	http.ListenAndServe(port, Log(http.DefaultServeMux))
}
