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

// 主结构
type D struct {
	files map[string]string
}

// 主
func main() {

	// 设置CPU核心数量
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 设置日志的结构
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds)

	// 运行
	NewD().Web()

}

// 创建对象
func NewD() *D {
	d := new(D)
	return d
}

// Web API 接口现场
func (d *D) Web() {
	// 建立监听
	if e := http.ListenAndServe(":8686", d); e != nil {
		panic(e)
	}
}

// Web API 的主接口方法
func (d *D) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 设置路由
	switch r.URL.Path {
	// 路由接口
	// ---------------------------
	case "/":
		Index(w, r)
	case "/ul":
		UL(w, r)
	case "dw":
		DW(w, r)
	}
}

// 上传文件接口
func UL(w http.ResponseWriter, r *http.Request) {
}

// 下载文件接口
func DW(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	// 获取文件名称
	fname := z.Trim(r.FormValue("file"))
	// 判断安装包是否存在
	if !z.Exists(fmt.Sprintf("files/%s", fname)) {
		w.WriteHeader(404)
		return
	}
	// 写入文件流
	z.FileRF(fmt.Sprintf("files/%s", fname), func(f *os.File) {
		_, err := io.Copy(w, bufio.NewReader(f))
		if err != nil {
			log.Println(err)
			return
		}
	})
}

// 主页，提供上传，搜索，列表
func Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("template/html/index.html")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, nil)
}
