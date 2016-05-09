package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var CONFIG *Config

// 主
func main() {

	// 设置CPU核心数量
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 设置日志的结构
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)

	// -------------------------------------------------------- //

	CONFIG = readConfig()

	// -------------------------------------------------------- //

	http.Handle("/css/", http.FileServer(http.Dir("template")))

	http.Handle("/js/", http.FileServer(http.Dir("template")))

	http.Handle("/files/", http.FileServer(http.Dir("template")))

	http.Handle("/images/", http.FileServer(http.Dir("template")))

	// -------------------------------------------------------- //

	http.HandleFunc("/", index)

	http.HandleFunc("/rmfile.go", rmfile)

	http.HandleFunc("/upload.go", upload)

	http.HandleFunc("/download.go", download)

	// -------------------------------------------------------- //

	// 建立监听
	if err := http.ListenAndServe(":8080", nil); err != nil {
		// 踢出错误
		log.Panic(err)
	}

}

type Size interface {
	Size() int64
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

		if sizeInterface, ok := file.(Size); ok {
			if float64(sizeInterface.Size()) > CONFIG.Size {
				http.Error(w, "超过文件大小限制", 500)
				return
			}
		}

		// 判断文件是否存在
		if Exists(fmt.Sprintf("files/%s", multi.Filename)) {
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

	}

	// 重定向
	http.Redirect(w, r, "/", http.StatusFound)

	// 返回
	return

}

// 下载文件接口
func download(w http.ResponseWriter, r *http.Request) {

	// 解析参数
	r.ParseForm()

	// 获取文件名称
	fname := Trim(r.FormValue("f"))

	// 添加头信息
	w.Header().Set("Content-Type", "multipart/form-data")

	// 添加头信息
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fname))

	// 判断安装包是否存在
	if !Exists(fmt.Sprintf("files/%s", fname)) {
		http.Error(w, fmt.Sprintf("WARN: [%s] file not exists ...", fname), 500)
		return
	}

	// 写入文件流
	FileRF(fmt.Sprintf("files/%s", fname), func(f *os.File) {
		_, err := io.Copy(w, bufio.NewReader(f))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})

	// 返回
	return

}

// 删除文件
func rmfile(w http.ResponseWriter, r *http.Request) {

	// cookie
	if _, err := r.Cookie("username"); err != nil {
		// 重定向
		http.Redirect(w, r, "/", http.StatusFound)
		// 返回
		return
	}

	// 解析参数
	r.ParseForm()

	// 获取文件名称
	fname := Trim(r.FormValue("f"))

	// 判断安装包是否存在
	if Exists(fmt.Sprintf("files/%s", fname)) && !IsBlank(fname) {
		// 删除
		Fremove(fmt.Sprintf("files/%s", fname))
	}

	// 重定向
	http.Redirect(w, r, "/", http.StatusFound)

	// 返回
	return

}

/*
	这里偷个懒
	应该将文件信息记录到数据库或者文件中
	我这个每次都去扫描，浪费资源
*/

type FileInfo struct {
	Id   int    // ID
	Name string // 文件名称
	Size string // 文件大小
	Date string // 上传日期
	Stat string // 权限状态
}

type Data struct {
	// 权限状态
	Stat string
	// 文件列表
	Files []*FileInfo
}

// 构造
func NewData() *Data {
	data := new(Data)
	data.Files = make([]*FileInfo, 0)
	return data
}

func index(w http.ResponseWriter, r *http.Request) {

	// 解析参数
	r.ParseForm()

	// 管理员
	var admin string

	// form
	if _, ok := r.Form[CONFIG.Admin]; ok {
		// cookie
		cookie := http.Cookie{Name: "username", Value: CONFIG.Admin, Expires: time.Now().Add(24 * time.Hour)}
		// cookie
		http.SetCookie(w, &cookie)
		// 管理员
		admin = CONFIG.Admin
	}

	// cookie
	if cookie, err := r.Cookie("username"); err == nil {
		// 权限
		if cookie.Value == CONFIG.Admin {
			// 管理员
			admin = cookie.Value
		}
	}

	// 获取文件名称
	fname := Trim(r.FormValue("f"))

	// 创建返回对象
	data := NewData()
	data.Stat = admin

	// ID
	var id int

	// 遍历本地文件
	filepath.Walk("files", func(ph string, f os.FileInfo, err error) error {
		// 文件不存在
		if f == nil {
			return nil
		}
		// 跳过文件夹
		if f.IsDir() {
			return nil
		}
		// 判断文件是否存在
		if IsBlank(fname) {
			// 累加
			id++
			// 记录文件
			data.Files = append(data.Files, &FileInfo{id, f.Name(), unitCapacity(f.Size()), f.ModTime().String(), admin})
		} else {
			// 检查包含
			if strings.Contains(strings.ToLower(f.Name()), strings.ToLower(fname)) {
				// 累加
				id++
				// 记录文件
				data.Files = append(data.Files, &FileInfo{id, f.Name(), unitCapacity(f.Size()), f.ModTime().String(), admin})
			}
		}
		// 返回
		return nil
	})

	// 解析主页面
	t, err := template.ParseFiles("template/default.html")
	if err != nil {
		// 输出错误信息
		http.Error(w, err.Error(), 500)
		return
	}

	// 执行
	t.Execute(w, data)

	// 返回
	return

}

func unitCapacity(size int64) string {
	if g := float64(size) / (1024 * 1024 * 1024); int64(g) > 0 {
		return fmt.Sprintf("%.2fG", g)
	} else if m := float64(size) / (1024 * 1024); int64(m) > 0 {
		return fmt.Sprintf("%.2fM", m)
	} else if k := float64(size) / (1024); int64(k) > 0 {
		return fmt.Sprintf("%.2fK", k)
	} else {
		return fmt.Sprintf("%dB", size)
	}
}

type Config struct {
	Size  float64 `json:"size"`
	Admin string  `json:"admin"`
}

func readConfig() *Config {
	// New ServerConf
	conf := new(Config)
	conf.Size = 1073741824
	conf.Admin = "admin"
	if !Exists("godw.conf") {
		log.Println("use default")
		log.Println("not found godw.conf")
		return conf
	}
	f, err := os.Open("godw.conf")
	if err != nil {
		log.Println("use default")
		log.Println(err.Error())
		return conf
	}
	bs, err := ioutil.ReadAll(bufio.NewReader(f))
	if err != nil {
		log.Println("use default")
		log.Println(err.Error())
		return conf
	}
	err = json.Unmarshal(bs, &conf)
	if err != nil {
		log.Println("use default")
		log.Println(err.Error())
		return conf
	}
	return conf
}

// 判断一个路径是否存在
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// 是不是空字符
func IsSpace(c byte) bool {
	if c >= 0x00 && c <= 0x20 {
		return true
	}
	return false
}

// 判断一个字符串是不是空白串，即（0x00 - 0x20 之内的字符均为空白字符）
func IsBlank(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if !IsSpace(b) {
			return false
		}
	}
	return true
}

// 去掉一个字符串左右的空白串，即（0x00 - 0x20 之内的字符均为空白字符）
// 与strings.TrimSpace功能一致
func Trim(s string) string {
	size := len(s)
	if size <= 0 {
		return s
	}
	l := 0
	for ; l < size; l++ {
		b := s[l]
		if !IsSpace(b) {
			break
		}
	}
	r := size - 1
	for ; r >= l; r-- {
		b := s[r]
		if !IsSpace(b) {
			break
		}
	}
	return string(s[l : r+1])
}

// Remove 文件
func Fremove(ph string) (err error) {
	err = os.Remove(ph)
	return err
}

/*
将从自己磁盘目录，只读的方式打开一个文件。如果文件不存在，或者打开错误，则返回 nil。
如果有错误，将打印 log

调用者将负责关闭文件
*/
func FileR(ph string) *os.File {
	f, err := os.Open(ph)
	if nil != err {
		return nil
	}
	return f
}

// 用回调的方式打文件以便读取内容，回调函数不需要关心文件关闭等问题
func FileRF(ph string, callback func(*os.File)) {
	f := FileR(ph)
	if nil != f {
		defer f.Close()
		callback(f)
	}
}
