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
	var concurrency = 20
	flag.IntVar(&concurrency, "c", 20, "Concurrency")
	proxy := flag.String("proxy", "", "HTTP proxy URL")
	flag.Parse()

	p := gahttp.NewPipelineWithClient(gahttp.NewClient(gahttp.SkipVerify))

	// Set the proxy if provided
	if *proxy != "" {
		proxyURL, err := url.Parse(*proxy)
		if err != nil {
			fmt.Printf("Failed to parse proxy URL: %v\n", err)
			return
		}
		p.Client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	p.SetConcurrency(concurrency)
	extractFn := gahttp.Wrap(extractTitle, gahttp.CloseBody)

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		p.Get(sc.Text(), extractFn)
	}
	p.Done()

	p.Wait()
}
