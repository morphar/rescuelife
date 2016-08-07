package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

const (
	DIE_RETRIES = 10
	DIE_SLEEP   = 10 * time.Second
)

type Crawler struct {
	http.Client
	UserAgent string
}

func NewCrawler() Crawler {
	jar, _ := cookiejar.New(nil)
	return Crawler{
		Client:    http.Client{Jar: jar},
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.1 Safari/537.36",
	}
}

func (c *Crawler) Get(u string) (res *http.Response, err error) {
	pUrl, _ := url.Parse(u)
	res, err = c.Client.Get(u)
	c.Client.Jar.SetCookies(pUrl, res.Cookies())
	return res, err
}

func (c *Crawler) Post(u string, bodyType string, body io.Reader) (res *http.Response, err error) {
	pUrl, _ := url.Parse(u)
	res, err = c.Client.Get(u)
	c.Client.Jar.SetCookies(pUrl, res.Cookies())
	return c.Client.Post(u, bodyType, body)
}

func (c *Crawler) PostForm(u string, data url.Values) (res *http.Response, err error) {
	pUrl, _ := url.Parse(u)
	res, err = c.Client.Get(u)
	c.Client.Jar.SetCookies(pUrl, res.Cookies())
	return c.Client.PostForm(u, data)
}

func (c *Crawler) GetOrDie(url string) (res *http.Response) {
	var err error

	fmt.Print("Fetching ", url, "...")

	for i := 0; i < DIE_RETRIES; i++ {
		res, err = c.Get(url)
		if err != nil || res.StatusCode != 200 {
			fmt.Print(".")
			time.Sleep(DIE_SLEEP)
		} else {
			break
		}
	}

	if err != nil {
		fmt.Println("\nFAILED! Picturelife is probably down, so try again later.")
		os.Exit(0)
	}
	fmt.Println(" Done!")
	return
}

func (c *Crawler) PostOrDie(url string, bodyType string, body io.Reader) (res *http.Response) {
	var err error

	for i := 0; i < DIE_RETRIES; i++ {
		res, err = c.Post(url, bodyType, body)
		if err != nil || res.StatusCode != 200 {
			time.Sleep(DIE_SLEEP)
		} else {
			break
		}
	}

	if err != nil {
		fmt.Println("\nFAILED! Picturelife is probably down, so try again later.")
		os.Exit(0)
	}
	return
}

func (c *Crawler) PostFormOrDie(url string, data url.Values) (res *http.Response) {
	var err error

	for i := 0; i < DIE_RETRIES; i++ {
		res, err = c.PostForm(url, data)
		if err != nil || res.StatusCode != 200 {
			fmt.Print(".")
			time.Sleep(DIE_SLEEP)
		} else {
			break
		}
	}

	if err != nil {
		fmt.Println("\nFAILED! Picturelife is probably down, so try again later.")
		os.Exit(0)
	}
	return
}
