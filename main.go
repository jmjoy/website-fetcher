package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func main() {
	// 参数
	dir := flag.String("d", ".", "指定文件存放的目录，默认是程序运行的目录")

	flag.Parse()

	// 设置当前目录
	os.Chdir(*dir)

	// 基础URL地址
	if flag.NArg() < 1 {
		log.Fatal("请指定要抓取的基础URL地址")
	}
	var err error
	BaseUrl, err = url.Parse(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	if BaseUrl.Host == "" || BaseUrl.Scheme == "" {
		log.Fatal("请指定正确的URL地址")
	}

	fetchAll(BaseUrl.String())

	Waiter.Wait()
}

func fetchAll(urls ...string) {
	if urls == nil {
		return
	}

	for i := range urls {
		Waiter.Add(1)
		go func(s string) {
			fetchAll(fetch(s)...)
			Waiter.Done()
		}(urls[i])
	}

}

func fetch(urlStr string) (subUrls []string) {

	// url处理
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Println(err)
		return
	}
	switch {
	case u.Host == "":
		u.Host = BaseUrl.Host
		u.Scheme = BaseUrl.Scheme

	case u.Host != BaseUrl.Host:
		return
	}

	// 网络链接
	resp, err := http.Get(u.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 判断是不是静态html文件
	cType := resp.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(cType, "text"):
		subUrls = fetchText(resp.Body, u)
		return
	default:
		fetchBlob(resp.Body, u)
		return
	}
}

func fetchBlob(reader io.Reader, u *url.URL) {
	// 获取相对路径
	filePath := u.Path
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}
	if filePath == "" {
		log.Println("不是有效的二进制文件路径")
		return
	}

	// 新建文件
	file, err := createFile(filePath)
	if err != nil {
		log.Println(err)
		return
	}

	// 写入文件
	_, err = io.Copy(file, reader)
	if err != nil {
		log.Println(err)
		return
	}

	// 加入到已访问行列
	addVisited(u)
}

func fetchText(reader io.Reader, u *url.URL) (subUrls []string) {

	// 读取
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println(err)
		return
	}

	// 网页所有内容
	body := string(buf)

	// 正则匹配超链接
	regs := []*regexp.Regexp{
		regexp.MustCompile(`[Hh][Rr][Ee][Ff]=["'](.*?)["']`),
		regexp.MustCompile(`[Ss][Rr][Cc]=["'](.*?)["']`),
		regexp.MustCompile(`[Uu][Rr][Ll]\(["']?(.*?)["']?\)`),
	}

	var matches [][]string
	for i := range regs {
		matches = append(matches, regs[i].FindAllStringSubmatch(body, -1)...)
	}

	// 获取相对路径
	filePath := dealSuffix(u.Path)

	// 换取本次URL的绝对文件路径
	basePath, err := filepath.Abs(path.Dir(filePath))
	if err != nil {
		log.Println(err)
		return
	}

	// 替换BODY的绝对路径
	subUrls = make([]string, 0, 64)
	for i := range matches {
		// 获取子URL
		str := matches[i][1]
		if !strings.HasPrefix(str, "http") && !strings.HasPrefix(str, "/") {
			continue
		}

		subU, err := url.Parse(str)
		if err != nil {
			log.Println(err)
			continue
		}

		if subU.Host != "" && subU.Host != BaseUrl.Host {
			continue
		}

		// 检查是否已经获取了
		if hasVisited(subU) {
			continue
		}

		subUrls = append(subUrls, str)

		replace := dealSuffix(subU.Path)

		// 获取绝对路径
		replace, err = filepath.Abs(replace)
		if err != nil {
			log.Println(err)
			continue
		}

		// 获取相对本页面的相对路径
		replace, err = filepath.Rel(basePath, replace)
		if err != nil {
			log.Println(err)
			continue
		}

		body = strings.Replace(body, str, replace, -1)
	}

	// 新建文件
	file, err := createFile(filePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	// 写入文件
	bodyReader := strings.NewReader(body)
	_, err = io.Copy(file, bodyReader)
	if err != nil {
		log.Println(err)
		return
	}

	// 加入到已访问行列
	addVisited(u)

	return
}

func dealSuffix(filePath string) string {
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	if filePath == "" {
		filePath = "index.html"
	}

	switch path.Ext(filePath) {
	case ".html", ".xhtml", ".shtml", ".css", ".js":
	case ".jpg", ".jpeg", ".png", ".bmp", ".gif", ".ico":
	case ".mp3", ".mp4", ".swf":
	default:
		filePath = filePath + ".html"
	}

	return filePath
}

func createFile(filePath string) (*os.File, error) {
	dir := path.Dir(filePath)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func addVisited(u *url.URL) {
	VisitLock.Lock()
	Visted = append(Visted, *u)
	VisitLock.Unlock()
}

func hasVisited(u *url.URL) bool {
	for _, v := range Visted {
		if u.Host != v.Host {
			continue
		}

		if u.Scheme != v.Scheme {
			continue
		}

		p1 := u.Path
		p2 := v.Path
		if strings.HasPrefix(p1, "/") {
			p1 = p1[1:]
		}
		if strings.HasPrefix(p2, "/") {
			p2 = p2[1:]
		}
		if p1 != p2 {
			continue
		}

		return true
	}
	return false
}

var (
	BaseUrl   *url.URL
	VisitLock sync.Mutex
	Visted    []url.URL
	Waiter    sync.WaitGroup
)
