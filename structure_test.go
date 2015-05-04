package main

/*
import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"testing"
)

func TestPath(t *testing.T) {
	for i, v := range urlList {
		u, _ := url.Parse(v)
		text := fmt.Sprintf("%d %v %s \n", i, v, u.Path)
		pathList := strings.Split(u.Path, "/")
		for _, s := range pathList {
			text = text + "--" + s + "-- "
		}
		text = text + "\n" + path.Join(pathList...)
		t.Log(text)
	}
}

func TestIsEqual(t *testing.T) {
	for _, v := range urlList {
		u, err := ParseURL(v, "www.baidu.com")
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
		u, _ := ParseURL(slist[i], "www.baidu.com")
		if u != ulist[i] {
			//t.Fatal("Error:", u, *ulist[i])
		}
	}

}

var urlList = []string{
	"http://www.baidu.com",
	"http://www.baidu.com/",
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
*/
