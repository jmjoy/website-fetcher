package main

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

type URL struct {
	*url.URL
	IsBlob   bool
	Filepath string
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

	if strings.HasPrefix(u.Path, "/") {
		u.Path = u.Path[1:]
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// 判断是不是静态html文件
	cType := resp.Header.Get("Content-Type")

	myU := &URL{URL: u, IsBlob: !strings.HasPrefix(cType, "text")}
	myU.filepath(host)

	return myU, nil
}

func (this *URL) IsEqual(other *URL) bool {
	if this.Path == other.Path && this.Host == other.Host && this.RawQuery == other.RawQuery {
		return false
	}
	return true
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

func (this *URL) filepath(host string) {
	this.Filepath = this.Path

	if this.Host != host {
		this.Filepath = "__other__/" + this.Host + "/" + this.Filepath
	}

	if this.IsBlob {
		return
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
