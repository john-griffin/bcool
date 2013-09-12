package main

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Item struct {
	Link        string `xml:"link"`
	Title       string `xml:"title"`
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

func OriginalFeedBody() []byte {
	res, err := http.Get("http://www.bleedingcool.com/feed/")
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
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "entry-content" {
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
	body := OriginalFeedBody()
	v := Rss{}
	err := xml.Unmarshal(body, &v)
	for key, value := range v.Channel.Items {
		value.Description = FetchFullDescription(value.Link)
		v.Channel.Items[key] = value
	}
	if err != nil {
		log.Fatal(err)
	}
	b, err := xml.Marshal(v)
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
