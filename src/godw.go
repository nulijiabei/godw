package main

import (
	"bufio"
	"fmt"
	z "github.com/nutzam/zgo"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
)

// 主
func main() {

	// 设置CPU核心数量
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 设置日志的结构
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds)

	// -------------------------------------------------------- //

	http.Handle("/css/", http.FileServer(http.Dir("template")))

	http.Handle("/js/", http.FileServer(http.Dir("template")))

	http.Handle("/files/", http.FileServer(http.Dir("template")))

	http.Handle("/images/", http.FileServer(http.Dir("template")))

	// -------------------------------------------------------- //

	http.HandleFunc("/", index)

	http.HandleFunc("/addfile.go", addfile)

	http.HandleFunc("/filelist.go", filelist)

	http.HandleFunc("/upload.go", upload)

	http.HandleFunc("/download.go", download)

	// -------------------------------------------------------- //

	// 建立监听
	if err := http.ListenAndServe(":8080", nil); err != nil {
		// 踢出错误
		log.Panic(err)
	}

}

// 上传文件接口
func upload(w http.ResponseWriter, r *http.Request) {

	// 加锁,写入
	if "POST" == r.Method {

		file, multi, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()

		// 判断文件是否存在
		if z.Exists(fmt.Sprintf("files/%s", multi.Filename)) {
			// 返回错误信息
			http.Error(w, fmt.Sprintf("WARN: [%s] file exists ...", multi.Filename), 500)
			return
		}

		f, err := os.Create(fmt.Sprintf("files/%s", multi.Filename))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		return

	}

}

// 下载文件接口
func download(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	// 获取文件名称
	fname := z.Trim(r.FormValue("f"))
	// 判断安装包是否存在
	if !z.Exists(fmt.Sprintf("files/%s", fname)) {
		http.Error(w, fmt.Sprintf("WARN: [%s] file not exists ...", fname), 500)
		return
	}
	// 写入文件流
	z.FileRF(fmt.Sprintf("files/%s", fname), func(f *os.File) {
		_, err := io.Copy(w, bufio.NewReader(f))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})
}

// 主页
func index(w http.ResponseWriter, r *http.Request) {
	// 解析主页面
	t, err := template.ParseFiles("template/index.html")
	if err != nil {
		// 输出错误信息
		http.Error(w, err.Error(), 500)
		return
	}
	// 执行
	t.Execute(w, nil)
}

func addfile(w http.ResponseWriter, r *http.Request) {
	// 解析主页面
	t, err := template.ParseFiles("template/files/addfile.html")
	if err != nil {
		// 输出错误信息
		http.Error(w, err.Error(), 500)
		return
	}
	// 执行
	t.Execute(w, nil)
}

func rmfile(w http.ResponseWriter, r *http.Request) {

}

func filelist(w http.ResponseWriter, r *http.Request) {
	// 解析主页面
	t, err := template.ParseFiles("template/files/filelist.html")
	if err != nil {
		// 输出错误信息
		http.Error(w, err.Error(), 500)
		return
	}
	// 执行
	t.Execute(w, nil)
}
