package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tomnomnom/gahttp"
	"golang.org/x/net/html"
)

func extractTitle(req *http.Request, resp *http.Response, err error) {
	if err != nil {
		return
	}

	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.StartTagToken && t.Data == "title" {
			if z.Next() == html.TextToken {
				title := strings.TrimSpace(z.Token().Data)
				fmt.Printf("%s (%s)\n", title, req.URL)
				break
			}
		}
	}
}

func main() {
	var (
		concurrency int
		proxyURL    string
	)

	flag.IntVar(&concurrency, "c", 20, "Concurrency")
	flag.StringVar(&proxyURL, "p", "", "Proxy URL (e.g., http://proxy.example.com:8080)")
	flag.Parse()

	p := gahttp.NewPipelineWithClient(newHTTPClient(proxyURL))
	p.SetConcurrency(concurrency)
	extractFn := gahttp.Wrap(extractTitle, gahttp.CloseBody)

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		p.Get(sc.Text(), extractFn)
	}
	p.Done()

	p.Wait()
}

func newHTTPClient(proxyURL string) *http.Client {
	if proxyURL == "" {
		return gahttp.NewClient(gahttp.SkipVerify)
	}

	proxyURLParsed, err := url.Parse(proxyURL)
	if err != nil {
		fmt.Printf("Failed to parse proxy URL: %v\n", err)
		os.Exit(1)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURLParsed),
	}
	return &http.Client{
		Transport: transport,
	}
}
