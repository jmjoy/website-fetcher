package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

type CType int

const (
	UNKNOW CType = iota
	IS_TEXT
	IS_BLOB
)

type URL struct {
	*url.URL
	url.Values
	PathList    []string
	filePath    string
	contentType CType
}

func ParseURL(rawurl, host string) (*URL, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("不支持http/https之外的协议")
	}

	if u.Host == "" {
		u.Host = host
	}

	uu := new(URL)
	uu.URL = u
	uu.Values = u.Query()
	uu.PathList = strings.Split(path.Clean(u.Path), "/")

	return uu, nil
}

func (this *URL) IsEqual(other *URL) bool {
	fmt.Println(*this)
	fmt.Println(*other)

	if this.Host != other.Host {
		return false
	}

	if len(this.PathList) != len(other.PathList) {
		return false
	}
	for i := range this.PathList {
		if this.PathList[i] != other.PathList[i] {
			return false
		}
	}

	for k, v1 := range this.Values {
		v2, ok := other.Values[k]
		if !ok {
			return false
		}
		if len(v1) != len(v2) {
			return false
		}
		for i := range v1 {
			if !strInSlice(v1[i], v2) {
				return false
			}
		}
	}

	return true
}

func (this *URL) IsTExt() (bool, error) {
	switch this.contentType {
	case UNKNOW:
		return false, errNeedGet
	case IS_TEXT:
		return true, nil
	case IS_BLOB:
		return false, nil
	default:
		return false, errors.New("不可能发生的错误")
	}
}

func (this *URL) Get() (*http.Response, error) {
	resp, err := http.Get(this.String())
	if err != nil {
		return nil, err
	}
	contentType := resp.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "text") {
		this.contentType = IS_TEXT
	} else {
		this.contentType = IS_BLOB
	}

	return resp, nil
}

func (this *URL) Relpath(basepath string) (string, error) {
	absBasepath, err := filepath.Abs(basepath)
	if err != nil {
		return "", err
	}

	absFilepath, err := filepath.Abs(this.filePath)
	if err != nil {
		return "", err
	}

	return filepath.Rel(absBasepath, absFilepath)
}

func (this *URL) Filepath(host string) (string, error) {
	if this.filePath != "" {
		return this.filePath, nil
	}

	lastPathIndex := len(this.PathList) - 1
	pathList := make([]string, len(this.PathList))
	copy(pathList, this.PathList)

	switch this.contentType {
	case UNKNOW:
		return "", errNeedGet

	case IS_TEXT:
		if pathList[lastPathIndex] == "" {
			pathList[lastPathIndex] = "__index__.html"
		} else {
			switch path.Ext(pathList[len(pathList)-1]) {
			case "html", "shtml", "xhtml", "css", "js":
			default:
				pathList[lastPathIndex] += ".html"
			}
		}

	case IS_BLOB:
		if pathList[lastPathIndex] == "" {
			pathList[lastPathIndex] = "__index__"
		}
	}

	if len(this.Values) != 0 {
		index := strings.LastIndex(pathList[lastPathIndex], ".")

		if index == -1 {
			pathList[lastPathIndex] = pathList[lastPathIndex] +
				"__" + this.Values.Encode() + "__"

		} else {
			pathList[lastPathIndex] = pathList[lastPathIndex][:index] +
				"__" + this.Values.Encode() + "__" +
				pathList[lastPathIndex][index:]
		}
	}

	var filePath string

	if this.Host == host {
		filePath = path.Join(pathList...)
	} else {
		newPathList := make([]string, len(pathList)+2)
		newPathList = append(newPathList, "__other__", host)
		newPathList = append(newPathList, pathList...)
		filePath = path.Join(newPathList...)
	}

	this.filePath = filePath

	return filePath, nil
}

func strInSlice(s string, slice []string) bool {
	for i := range slice {
		if s == slice[i] {
			return true
		}
	}
	return false
}

var (
	errNeedGet = errors.New("请先调用Get")
)
