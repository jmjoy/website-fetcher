package main

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

// 增强URL
type URL struct {
	*url.URL
	FilePath string // 生成的静态文件相对于工作目录的路径
}

// ParseURL 将s转化为增强URL，host为默认的域名（如果s没有指定域名）
func ParseURL(s, host string) (*URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	// 默认协议头为http，如果不是http(s)则返回错误
	switch {
	case u.Scheme == "":
		u.Scheme = "http"

	case u.Scheme != "http" && u.Scheme != "https":
		return nil, ErrNotHttp
	}

	// 默认域名
	if u.Host == "" {
		u.Host = host
	}

	mu := new(URL)

	mu.URL = u

	pathList := strings.Split(mu.Path, "/")

	// 如果url路径没有后缀或为空，Filepath有一个默认的文件名
	if lastIndex := len(pathList) - 1; lastIndex > 0 {
		if pathList[lastIndex] == "" {
			pathList[lastIndex] = "__index__"
		}
	} else {
		pathList = append(pathList, "__index__")
	}

	// 在Filepath的文件名里加上Query参数
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

	// 组合Filepath，格式类似于 域名/xxx/.../xxx[.xxx]
	newPathList := make([]string, 0, len(pathList)+2)
	newPathList = append(newPathList, u.Host)
	newPathList = append(newPathList, pathList...)

	mu.FilePath = path.Join(newPathList...)

	return mu, nil
}

// Get 通过http的get请求url获取Response和是否文本数据（否则是二进制）
func (this *URL) Get() (resp *http.Response, isText bool, err error) {
	resp, err = http.Get(this.String())
	if err != nil {
		return
	}

	cType := resp.Header.Get("Content-Type")
	isText = strings.HasPrefix(cType, "text")

	return
}

// Equal 通过比较两个URL的Host和Filepath判断是否相等
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
