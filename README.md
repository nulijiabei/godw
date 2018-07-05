----------------
Godw

<a href="https://godoc.org/github.com/nulijiabei/godw"><img src="https://godoc.org/github.com/nulijiabei/godw?status.svg" alt="GoDoc"></a>

网络文件管理
	
	通过一个简单的页面，来将重要的文件存储到云端 ...
	通过一个简单的页面，来方便讲云端文件下载本地 ...

[PATH]

    [godw]     主程序
    [files]    存储目录
    [template] 模版文件

注意：主程序需要在[PATH]目录下执行,因为程序内部使用的相对目录,当然,你也可以修改为绝对目录

因为偷懒，所以并没有讲目录文件信息记录到数据库或者记录文件内，而是每次遍历，非常浪费资源

如果程序在环境并发很高的话，建议修改记录到数据库等

[UPDATE]

	增加权限管理 [session]
	
	普通用户: 查看，下载，无法删除操作
	
	管理员: 添加，查看，下载，删除，均可

[ 管理员 OR 普通用户 ]

	http://127.0.0.1:8080
	http://127.0.0.1:8080/?admin [这里的admin在godw.conf中设置]

[ Linux Bash 上传 ]

	curl -F "file=@a.jpg;filename=a.jpg"  http:/xxx.xxx.com:8080/upload
	curl -F "file=@a.jpg;filename=a.jpg"  http:/xxx.xxx.com:8080/upload/f

----------------

![image](https://raw.githubusercontent.com/nulijiabei/godw/master/screenshot.png)


