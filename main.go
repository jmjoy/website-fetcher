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
	"regexp"
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
	handleBaseURL(flag.Arg(0))

	// 新建文件夹
	handleDir()

	// 执行抓取
	ToVisit.PushBack(*BaseURL)
	for {
		elem := ToVisit.Front()
		if elem == nil {
			break
		}
		u := ToVisit.Remove(elem).(URL)
		fetch(&u)
		Visited.PushBack(u)
	}

	log.Printf("共抓取 %d 次，其中静态页面 %d 次，其他资源 %d 次\n", TextCount+BlobCount, TextCount, BlobCount)
}

func fetch(u *URL) {
	// 网络链接
	resp, err := http.Get(u.String())
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	// 判断是不是静态html文件
	if u.IsBlob {
		fetchBlob(resp.Body, u)
	} else {
		fetchText(resp.Body, u)
	}

	log.Println(u.String())
}

func fetchBlob(reader io.Reader, u *URL) {
	// 获取相对路径
	if u.Filepath == "" {
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

func fetchText(reader io.Reader, u *URL) {
	// 新建文件
	file, err := createFile(u.Filepath)
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

			subU, err := ParseURL(link, BaseURL.Host)
			if err != nil {
				log.Println(err)
				continue
			}

			handlePush(subU)

			// 修改body中的路径
			relP, err := subU.Relpath(path.Dir(u.Filepath))
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

func handlePush(u *URL) {
	for e := Visited.Front(); e != nil; e = e.Next() {
		nu := e.Value.(URL)
		if u.IsEqual(&nu) {
			return
		}
	}

	for e := ToVisit.Front(); e != nil; e = e.Next() {
		nu := e.Value.(URL)
		if u.IsEqual(&nu) {
			return
		}
	}

	ToVisit.PushBack(*u)
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

func handleBaseURL(s string) {
	// 基础URL地址
	u, err := url.Parse(s)

	if err != nil || u.Host == "" || u.Scheme == "" {
		log.Fatal("请指定指定完整正确的URL地址")
	}

	BaseURL, err = ParseURL(s, u.Host)
	if err != nil {
		log.Fatal(err)
	}

	if BaseURL.IsBlob {
		log.Fatal("URL地址获取的内容不是文本类型")
	}

	if IsAll {
		newURL := new(URL)
		newURL.Scheme = BaseURL.Scheme
		newURL.Host = BaseURL.Host
		newURL.IsBlob = false
		newURL.Filepath = "__index__.html"
		BaseURL = newURL
	}
}

func handleDir() {
	if BaseDir == "" {
		BaseDir = BaseURL.Host
	}

	// 创建目录
	err := os.Mkdir(BaseDir, 0777)
	if err != nil {
		log.Fatal(err)
	}

	// 设置当前目录
	os.Chdir(BaseDir)
	if err != nil {
		log.Fatal(err)
	}
}

func printHelp() {
	fmt.Println(`website-fetcher 适用于抓取文档类型网站的小工具`)
	fmt.Println(`用法: website-fetcher [-dir|-count|-deepth|-all|-help] URL`)
	fmt.Println()
	flag.PrintDefaults()
}

var (
	BaseDir  string
	DirLimit int
	Deepth   int
	IsAll    bool
	IsHelp   bool

	BaseURL *URL

	TextCount int
	BlobCount int

	ToVisit = list.New()
	Visited = list.New()
)

var Regs = []*regexp.Regexp{
	regexp.MustCompile(`(?i)href="(.*?)"`),
	regexp.MustCompile(`(?i)href='(.*?)'`),
	regexp.MustCompile(`(?i)src="(.*?)"`),
	regexp.MustCompile(`(?i)src='(.*?)'`),
	regexp.MustCompile(`(?i)url\((.*?)\)`),
	regexp.MustCompile(`(?i)url\("?(.*?)"?\)`),
	regexp.MustCompile(`(?i)url\('?(.*?)'?\)`),
}
