package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func init() {
	// 输入参数
	flag.StringVar(&BaseDir, "dir", "", "存放网页文件的基本目录，默认为网站域名")
	flag.IntVar(&DirLimit, "count", 255, "限制每个文件夹的文件数，默认255个")
	flag.IntVar(&Deepth, "deepth", 16, "限制文件的深度，默认16层")
	flag.BoolVar(&IsAll, "all", false, "是否要抓取整个网站，默认只抓取指定URL以下的网页")
	flag.BoolVar(&IsHelp, "help", false, "获取帮助")

	flag.Parse()
}

func main() {
	// 判断是不是要获取帮助信息
	if IsHelp || flag.NArg() < 1 {
		printHelp()
		return
	}

	// 基础URL地址
	baseUrl, err := ParseDetailURL(flag.Arg(0))

	if err != nil || baseUrl.Host == "" || baseUrl.Scheme == "" {
		log.Fatal("请指定正确的URL地址")
	}

	BaseHost = baseUrl.Host

	if BaseDir == "" {
		BaseDir = BaseHost
	}

	// 创建目录
	err = os.Mkdir(BaseDir, 0777)
	if err != nil {
		log.Fatal(err)
	}

	// 设置当前目录
	os.Chdir(BaseDir)
	if err != nil {
		log.Fatal(err)
	}

	// 执行抓取
	ToVisit.PushBack(*baseUrl)
	for {
		elem := ToVisit.Front()
		if elem == nil {
			break
		}
		u := ToVisit.Remove(elem).(DetailURL)
		fetch(u)
		Visted.PushBack(u)
	}

	log.Printf("共抓取 %d 次，其中静态页面 %d 次，其他资源 %d 次\n", TextCount+BlobCount, TextCount, BlobCount)
}

func fetch(u DetailURL) {
	// 判断是否已经访问过
	if hasVisited(&u) {
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
		fetchText(resp.Body, &u)
	default:
		fetchBlob(resp.Body, &u)
	}
	log.Println(u.String())
}

func fetchBlob(reader io.Reader, u *url.URL) {
	// 获取相对路径
	if u.Path == "" {
		log.Println(u.Path, ": 不是有效的二进制文件路径")
		return
	}

	// 新建文件
	file, err := createFile(u.Path)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, reader)
	if err != nil {
		log.Println(err)
		return
	}

	BlobCount++
}

func fetchText(reader io.Reader, u *url.URL) {
	// 换取本次URL的绝对文件路径
	basePath := dealSuffix(u.Path)
	baseDir, err := filepath.Abs(path.Dir(basePath))
	if err != nil {
		log.Println(err)
		return
	}

	// 新建文件
	file, err := createFile(basePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	// ReadLine 读取BODY
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		// 正则匹配超链接
		var matches [][]int

		for i := range Regs {
			matches = append(matches, Regs[i].FindAllStringSubmatchIndex(line, -1)...)
		}

		newLine := line
		for _, v := range matches {
			// 获取子URL
			link := line[v[2]:v[3]]
			if !strings.HasPrefix(link, "http") && !strings.HasPrefix(link, "/") {
				continue
			}

			subU, err := url.Parse(link)
			if err != nil {
				log.Println(err)
				continue
			}

			if subU.Host == "" {
				subU.Host = BaseHost
				subU.Scheme = "http"
			}

			if subU.Scheme == "" {
				subU.Scheme = "http"
			}

			if subU.Host != BaseHost {
				continue
			}

			if strings.HasPrefix(subU.Path, "/") {
				subU.Path = subU.Path[1:]
			}

			subU = cleanURL(subU)

			if !hasVisited(subU) {
				ToVisit.PushBack(*subU)
			}

			// 修改body中的路径
			relP, err := filepath.Abs(dealSuffix(subU.Path))
			if err != nil {
				log.Println(err)
				continue
			}
			relP, err = filepath.Rel(baseDir, relP)
			if err != nil {
				log.Println(err)
				continue
			}

			newLine = line[:v[2]] + relP + line[v[3]:]
		}

		// 修改后并写入文件
		_, err = file.WriteString(newLine + "\n")
		if err != nil {
			log.Println(err)
		}
	}

	if err = scanner.Err(); err != nil {
		log.Println(err)
		return
	}

	TextCount++
}

func dealSuffix(filePath string) string {
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
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

func hasVisited(u *url.URL) bool {
	for e := Visted.Front(); e != nil; e = e.Next() {
		if *u == e.Value {
			return true
		}
	}
	return false
}

func cleanURL(u *url.URL) *url.URL {
	return &url.URL{
		Host:   u.Host,
		Path:   u.Path,
		Scheme: u.Scheme,
	}
}

func printHelp() {
	fmt.Println(`website-fetcher 适用于抓取文档类型网站的小工具`)
	fmt.Println(`用法: website-fetcher [-dir|-count|-deepth|-all|-help] URL`)
	fmt.Println()
	flag.PrintDefaults()
}

var (
	TextCount int
	BlobCount int
)

var BaseHost string

var (
	ToVisit = list.New()
	Visted  Dir
)

var Regs = []*regexp.Regexp{
	regexp.MustCompile(`[Hh][Rr][Ee][Ff]=["'](.*?)["']`),
	regexp.MustCompile(`[Ss][Rr][Cc]=["'](.*?)["']`),
	regexp.MustCompile(`[Uu][Rr][Ll]\(["']?(.*?)["']?\)`),
}

var (
	BaseDir  string
	DirLimit int
	Deepth   int
	IsAll    bool
	IsHelp   bool
)
