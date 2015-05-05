package main

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

type URL struct {
	*url.URL
	FilePath string
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

	pathList := strings.Split(mu.Path, "/")

	if lastIndex := len(pathList) - 1; lastIndex > 0 {
		if pathList[lastIndex] == "" {
			pathList[lastIndex] = "__index__"
		}
	} else {
		pathList = append(pathList, "__index__")
	}

	values := u.Query()
	if values != nil && len(values) > 0 {
		for k := range values {
			sort.Strings(values[k])
		}
		lastIndex := len(pathList) - 1

		if lastPos := strings.LastIndex(pathList[lastIndex], "."); lastPos != -1 {
			pathList[lastIndex] = pathList[lastIndex][:lastPos] +
				"__" + values.Encode() + "__" +
				pathList[lastIndex][lastPos:]
		} else {
			pathList[lastIndex] = "__" + values.Encode() + "__"
		}
	}

	newPathList := make([]string, 0, len(pathList)+2)
	newPathList = append(newPathList, u.Host)
	newPathList = append(newPathList, pathList...)

	mu.FilePath = path.Join(newPathList...)

	return mu, nil
}

func (this *URL) Get() (resp *http.Response, isText bool, err error) {
	resp, err = http.Get(this.String())
	if err != nil {
		return
	}

	cType := resp.Header.Get("Content-Type")
	isText = strings.HasPrefix(cType, "text")

	return
}

func (this *URL) Equal(other *URL) bool {

	if this.Host != other.Host {
		return false
	}

	if this.FilePath != other.FilePath {
		return false
	}

	return true
}

var (
	ErrNotHttp = errors.New("不支持http/https之外的协议")
)
