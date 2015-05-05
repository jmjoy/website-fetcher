package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestMe(t *testing.T) {
	t.Log(strings.Split("", "/"))
}

func TestParseURL(t *testing.T) {
	for i := range strlist {
		u, err := ParseURL(strlist[i], "www.baidu.com")
		if err != nil {
			t.Fatal(err)
		}
		if !u.Equal(&urlist[i]) {
			t.Fatalf("%#v\n%#v", *u, urlist[i])
		}
	}
}

var (
	strlist = [...]string{
		"http://www.baidu.com",
		"http://www.baidu.com/index.html",
		"/a/b/c/index.html",
		"/a/b/c/index.html?a=1&a=2&b=3",
		"http://www.google.com/a/b/c/",
	}

	urlist = [...]URL{
		URL{
			URL: &url.URL{
				Host:   "www.baidu.com",
				Scheme: "http",
			},
			FilePath: "www.baidu.com/__index__",
		},
		URL{
			URL: &url.URL{
				Host:   "www.baidu.com",
				Scheme: "http",
			},
			FilePath: "www.baidu.com/index.html",
		},
		URL{
			URL: &url.URL{
				Host:   "www.baidu.com",
				Scheme: "http",
			},
			FilePath: "www.baidu.com/a/b/c/index.html",
		},
		URL{
			URL: &url.URL{
				Host:   "www.baidu.com",
				Scheme: "http",
			},
			FilePath: "www.baidu.com/a/b/c/index__a=1&a=2&b=3__.html",
		},
		URL{
			URL: &url.URL{
				Host:   "www.google.com",
				Scheme: "http",
			},
			FilePath: "www.google.com/a/b/c/__index__",
		},
	}
)
