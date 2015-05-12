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
	// TODO 未实现
	// flag.IntVar(&DirLimit, "count", 1000, "限制每个文件夹的文件数，默认1000个")
	flag.IntVar(&Deepth, "deepth", 16, "限制文件的深度，默认16层")
	flag.BoolVar(&IsAll, "all", false, "是否要抓取整个网站，默认只抓取指定URL以下的网页")
	flag.BoolVar(&IsHelp, "help", false, "获取帮助")
	// TODO 未实现URL矫正功能
	// flag.BoolVar(&IsRaw, "raw", false, "是否原样抓取网页内容，不进行URL矫正")

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

	// 对本域名下的文本数据执行crawl，二进制或其他域名的只下载
	if isText && u.Host == BaseURL.Host {
		fetchText(file, resp.Body, u)
	} else {
		fetchBlob(file, resp.Body, u)
	}

	log.Println(u.String())
}

// 获取二进制类型或其他域名的数据
func fetchBlob(file io.Writer, reader io.Reader, u *URL) {
	// 写入文件
	_, err := io.Copy(file, reader)
	if err != nil {
		log.Println(err)
		return
	}

	BlobCount++
}

// 获取文本数据，并crawl子url
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

			// 看看子url符不符合抓取条件
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
	// 判断深度，-1是减去Filepath最前面的域名
	if len(strings.Split(u.FilePath, string(os.PathSeparator)))-1 > Deepth {
		return
	}

	// 判断是不是基础URL底下的url
	if BasePath != "" {
		if u.Host == BaseURL.Host && !strings.HasPrefix(u.FilePath, BasePath) {
			return
		}
	}

	// 判断是不是已经抓取过了
	for e := Visited.Front(); e != nil; e = e.Next() {
		nu := e.Value.(URL)
		if u.Equal(&nu) {
			return
		}
	}

	// 判断是不是已经在待抓取列表里了
	for e := ToVisit.Front(); e != nil; e = e.Next() {
		nu := e.Value.(URL)
		if u.Equal(&nu) {
			return
		}
	}

	ToVisit.PushBack(*u)
}

func createFile(filePath string) (*os.File, error) {
	// 新建需要的文件夹
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

// 处理基础URL地址
func handleBaseURL(s string) {
	u, err := url.Parse(s)

	if err != nil || u.Host == "" || u.Scheme == "" {
		log.Fatal("请指定指定完整正确的URL地址")
	}

	BaseURL, err = ParseURL(s, u.Host)
	if err != nil {
		log.Fatal(err)
	}

	// 用于判断是否子url
	if !IsAll {
		BasePath = path.Dir(BaseURL.FilePath)
	}

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
	BaseDir  string // 工作目录
	DirLimit int    // 限制一个文件夹下可以有多少文件
	Deepth   int    // 限制文件相对工作目录的深度
	IsAll    bool   // 是否抓取整个网站，不只抓取基础地址下的URL
	IsHelp   bool   // 是否只打印帮助信息
	IsRaw    bool   // 是否保留文件中原始url，不转成本地可用形式

	BaseURL  *URL   // 基础地址URL
	BasePath string // 用于判断基础地址下的URL

	TextCount int // 已下载文本数据统计数量
	BlobCount int // 已下载二进制数据统计数量

	ToVisit = list.New() // 待抓取URL列表
	Visited = list.New() // 已抓取URL列表
)

var Regs = []*regexp.Regexp{
	regexp.MustCompile(`(?i)href=(["'])(.*?)(["'])`),
	regexp.MustCompile(`(?i)src=(["'])(.*?)(["'])`),
	regexp.MustCompile(`(?i)url\((["']?)'(.*?)(["']?)\)`),
}
