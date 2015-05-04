package main

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type URL struct {
	*url.URL
	url.Values
	PathList []string
}

func ParseURL(s, host string) (*URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	switch {
	case u.Scheme == "":
		u.Scheme = "http"

	case u.Scheme != "http" && u.Scheme != "https":
		return nil, ErrNotHttp
	}

	if u.Host == "" {
		u.Host = host
	}

	mu := new(URL)
	mu.URL = u
	mu.Values = u.Query()
	mu.PathList = strings.Split(u.Path, "/")

	return mu, nil
}

func (this *URL) Get() (*http.Response, error) {
	return http.Get(this.String())
}

func (this *URL) FilePath() string {

	pathList := make([]string, 0, len(this.PathList)+1)
	pathList = append(pathList, this.Host)
	pathList = append(pathList, this.PathList...)

	lastIndex := len(pathList) - 1
	if pathList[lastIndex] == "" {
		pathList[lastIndex] = "__index__.html"
	}

	return path.Join(pathList...)
}

var (
	ErrNotHttp = errors.New("不支持http/https之外的协议")
)
