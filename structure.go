package main

import (
	"errors"
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

	absFilepath, err := filepath.Abs(this.Filepath)
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
	copy()

	switch this.contentType {
	case UNKNOW:
		return "", errNeedGet

	case IS_TEXT:
		if this.PathList[lastPathIndex] == "" {
			this.PathList[lastPathIndex] = "__index__.html"
		} else {
			switch path.Ext(this.PathList[len(this.PathList)-1]) {
			case "html", "shtml", "xhtml", "css", "js":
			default:
				this.PathList[lastPathIndex] += ".html"
			}
		}

	case IS_BLOB:
		if this.PathList[lastPathIndex] == "" {
			this.PathList[lastPathIndex] = "__index__"
		}
	}

	if len(this.Values) != 0 {
		index := strings.LastIndex(this.PathList[lastPathIndex], ".")

		if index == -1 {
			this.PathList[lastPathIndex] = this.PathList[lastPathIndex] +
				"__" + this.Values.Encode() + "__"

		} else {
			this.PathList[lastPathIndex] = this.PathList[lastPathIndex][:index] +
				"__" + this.Values.Encode() + "__" +
				this.PathList[lastPathIndex][index:]
		}
	}

	if this.Host != host {
		this.Filepath = "__other__/" + this.Host + "/" + this.Filepath
	}

	if this.Filepath == "" || strings.HasSuffix(this.Filepath, "/") {
		this.Filepath += "__index__.html"
	}

	switch path.Ext(this.Filepath) {
	case ".html", ".xhtml", ".shtml", ".css", ".js":
	default:
		this.Filepath += ".html"
	}

	if this.RawQuery != "" {
		lastIndex := strings.LastIndex(this.Filepath, ".")
		this.Filepath = this.Filepath[:lastIndex] + "__" + this.RawQuery + "__" + this.Filepath[lastIndex:]
	}
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
