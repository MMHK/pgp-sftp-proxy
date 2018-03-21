# pgp-sftp-proxy
sftp http proxy with PGP encrypt


具体实现的功能如下:

- PGP加密文件
- 上传文件并PGP加密， 到Zurich的sftp
- 自动识别图片文件，将图片文件转换成PDF后再PGP加密上传
- 自带http server，使用http rest API操作
- API文档请编译后执行 `http://127.0.0.1:3333/sample/api.html`

外部依赖：
- [wkhtmltopdf](http://wkhtmltopdf.org/) 用于将html文件转换成PDF

## wkhtmltopdf  安装
----
`*nix` 下不要使用 `APT` / `YUM` 安装，官方有预编译二进制版本。安装版本需要`Xorg`支持。

由于wkhtmltopdf是使用系统自带的字体渲染HTML页面的，所以请预先将使用到的Font安装到目标系统中。

## 编译
----
- 安装Golang环境, Go >= 1.5
- checkout 源码
- 在源码目录 执行` go get -v `签出所有的依赖库
- ` go build ` 编译成二进制可执行文件
- 执行文件 ` GOPGP -c ./config.json`

## 配置文件
----
该项目使用json文件进行配置，具体例子如下
``
{
    "listen": "127.0.0.1:3333",
    "tmp_path": "./web_root/temp",
	"web_root" : "./web_root",
	"ssh" : {
		"host" : "",
		"user" : "",
		"password" : "",
		"key" : ""
	},
	"deploy_path" : {
		"dev" : "/Interface_Development_Files/",
		"pro" : "/Interface_Production_Files/",
		"test" : "/Interface_UAT_Files/"
	}
}
``

- `listen` 启动http service时绑定的地址
- `tmp_path` 临时文件的保存路径，一般临时包括：上传图片的原图、待上传到Zurich的文件、待转换的HTML文件，
  这些文件一般会在使用后马上删除，不过也不排除程序问题没有删除的文件。
- `web_root` http service使用的webroot
- `ssh` Zurich sftp的相关登录信息
- `deploy_path`  Zurich sftp的发布路径，用于区分不同的运行环境，一般不用更改