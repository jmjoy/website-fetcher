package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

func init() {
	// 输入参数
	flag.StringVar(&BaseDir, "dir", "web", "存放网页文件的基本目录，默认为web")
	flag.IntVar(&DirLimit, "count", 255, "限制每个文件夹的文件数，默认255个")
	flag.IntVar(&Deepth, "deepth", 16, "限制文件的深度，默认16层")
	flag.BoolVar(&IsAll, "all", false, "是否要抓取整个网站，默认只抓取指定URL以下的网页")
	flag.BoolVar(&IsHelp, "help", false, "获取帮助")
	flag.BoolVar(&IsRaw, "raw", false, "是否原样抓取网页内容，不进行URL矫正")

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
	resp, isText, err := u.Get()
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	// 新建文件
	file, err := createFile(u.FilePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	// 不同类型文件不同处理方式
	if isText && u.Host == BaseURL.Host {
		fetchText(file, resp.Body, u)
	} else {
		fetchBlob(file, resp.Body, u)
	}

	log.Println(u.String())
}

func fetchBlob(file io.Writer, reader io.Reader, u *URL) {
	// 写入文件
	_, err := io.Copy(file, reader)
	if err != nil {
		log.Println(err)
		return
	}

	BlobCount++
}

func fetchText(file io.Writer, reader io.Reader, u *URL) {
	// ReadLine 读取BODY
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		// 正则匹配超链接
		var matches [][]int

		for i := range Regs {
			matches = append(matches, Regs[i].FindAllStringSubmatchIndex(line, -1)...)
		}

		//newLine := line
		for _, v := range matches {
			// 左右引号不对称的情况，因为go的正则不支持向后引用，所以这样写
			if line[v[2]:v[3]] != line[v[6]:v[7]] {
				continue
			}

			// 获取子URL
			link := line[v[4]:v[5]]

			subU, err := ParseURL(link, BaseURL.Host)
			if err != nil {
				log.Println(err)
				continue
			}

			handlePush(subU)

			// 修改body中的路径
			//relP, err := subU.Relpath(path.Dir(u.Filepath))
			//if err != nil {
			//    log.Println(err)
			//    continue
			//}

			//newLine = line[:v[2]] + relP + line[v[3]:]
		}

		// 修改后并写入文件
		_, err := io.WriteString(file, line+"\n")
		if err != nil {
			log.Println(err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
		return
	}

	TextCount++
}

func handlePush(u *URL) {
	if BasePath != "" {
		if u.Host == BaseURL.Host && !strings.HasPrefix(u.FilePath, BasePath) {
			return
		}
	}

	for e := Visited.Front(); e != nil; e = e.Next() {
		nu := e.Value.(URL)
		if u.Equal(&nu) {
			return
		}
	}

	for e := ToVisit.Front(); e != nil; e = e.Next() {
		nu := e.Value.(URL)
		if u.Equal(&nu) {
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

	if !IsAll {
		BasePath = path.Dir(BaseURL.FilePath)
	}

	//if BaseURL.IsBlob {
	//    log.Fatal("URL地址获取的内容不是文本类型")
	//}

	//if IsAll {
	//    newURL := new(URL)
	//    newURL.Scheme = BaseURL.Scheme
	//    newURL.Host = BaseURL.Host
	//    newURL.IsBlob = false
	//    newURL.Filepath = "__index__.html"
	//    BaseURL = newURL
	//}
}

func handleDir() {
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
	IsRaw    bool

	BaseURL  *URL
	BasePath string

	TextCount int
	BlobCount int

	ToVisit = list.New()
	Visited = list.New()
)

var Regs = []*regexp.Regexp{
	regexp.MustCompile(`(?i)href=(["'])(.*?)(["'])`),
	regexp.MustCompile(`(?i)src=(["'])(.*?)(["'])`),
	regexp.MustCompile(`(?i)url\((["']?)'(.*?)(["']?)\)`),
}
