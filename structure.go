package main

import (
	"net/url"
	"path/filepath"
	"strings"
)

type DetailURL struct {
	*url.URL
	PathList []string
	url.Values
}

type Dir struct {
	Name   string
	Deepth int
	Urls   []*DetailURL
	Dirs   []*Dir
}

func ParseDetailURL(rawurl string) (*DetailURL, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	du := new(DetailURL)
	du.URL = u

	path := filepath.Clean(u.Path)
	pathList := strings.Split(path, "/")

	for _, item := range pathList {
		if item != "" {
			du.PathList = append(du.PathList, item)
		}
	}

	du.Values = u.Query()

	return du, nil
}

func InDir(d Dir, u *DetailURL) bool {
	for _, pathname := range u.PathList[:len(u.PathList)-1] {
		for j := range d.Dirs {
			if pathname == d.Dirs[j].Name {
				*d = d.Dirs[j]
			}
		}
		return false
	}
}
