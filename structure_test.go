package main

import (
	"net/url"
	"testing"
)

func TestIsEqual(t *testing.T) {
	for _, v := range urlList {
		u, err := ParseURL(v, "www.baidu.com", false)
		if err != nil {
			t.Log(err)
		} else {
			t.Log(*u)
			t.Logf("###%s###\n", u.Filepath)
		}
	}
}

func TestParseURL(t *testing.T) {
	ulist := []*URL{
		&URL{
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.baidu.com",
			},
			IsBlob:   false,
			Filepath: "__index__.html",
		},
	}

	slist := []string{
		"http://www.baidu.com",
	}

	for i := 0; i < len(ulist); i++ {
		u, _ := ParseURL(slist[i], "www.baidu.com", false)
		if u != ulist[i] {
			//t.Fatal("Error:", u, *ulist[i])
		}
	}

}

var urlList = []string{
	"http://www.baidu.com",
	"http://www.baidu.com/index.html",
	"/index.html",
	"http://www.baidu.com/index.css",
	"/index.css",
	"http://www.baidu.com/a/b/index.js",
	"/a/b/index.js",
	"http://www.baidu.com/a/b/../c/index.do",
	"/a/b/../c/index.do",
	"http://www.google.com/a/b/c/index.html",
}
