# website-fetcher

简单的静态网站抓取器，可以用来抓取文档类型的网站

## 安装

        $ go get github.com/jmjoy/website-fetcher

## 用法

### 最简单的

        $ website-fetcher [URL]

        比如下载W3C的教程：

        $ website-fetcher http://www.w3school.com.cn/

### 帮助

        $ website-fetcher

        website-fetcher 适用于抓取文档类型网站的小工具
        用法: website-fetcher [-dir|-count|-deepth|-all|-help] URL

          -all=false: 是否要抓取整个网站，默认只抓取指定URL以下的网页
          -deepth=16: 限制文件的深度，默认16层
          -dir="web": 存放网页文件的基本目录，默认为web
          -help=false: 获取帮助

## TODO

- [ ] Directory files limit
- [ ] Raw Option

## License

YOU CAN DO WHAT THE FUCK YOU WANT TO DO LICENSE
