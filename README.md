# pgp-sftp-proxy
sftp http proxy with PGP encrypt


具体实现的功能如下:

- PGP加密文件
- 上传文件并PGP加密， 到Zurich的sftp
- 自动识别图片文件，将图片文件转换成PDF后再PGP加密上传
- 自带http server，使用http rest API操作
- API文档请编译后执行 `http://127.0.0.1:3333/swagger/index.html`

外部依赖：
- [gopdf](https://github.com/signintech/gopdf) 用于将图片文件转换成PDF

## 编译
----
- 安装Golang环境, Go >= 1.12
- checkout 源码
- 在源码目录 执行` go mod vendor `签出所有的依赖库
- ` go build -o pgp-sftp-proxy . ` 编译成二进制可执行文件
- 执行文件 ` pgp-sftp-proxy -c ./config.json`

## 配置文件
----
该项目使用json文件进行配置，具体例子如下

```JS
{
    "listen": "127.0.0.1:3333", //服务绑定的host
    "tmp_path": "./web_root/temp", //文件缓存目录
	"web_root" : "./web_root", //API 文档说明目录
	"ssh" : {
		"host" : "", //ssh 远程登录host
		"user" : "", //ssh 远程登录账户
		"password" : "", //ssh 远程登录密码
		"key" : "" //ssh 远程登录密匙
	},
	"deploy_path" : {
		"dev" : "/Interface_Development_Files/", //sftp 远程开发目录文件夹
		"pro" : "/Interface_Production_Files/", //sftp 远程产品目录文件夹
		"test" : "/Interface_UAT_Files/" //sftp 远程测试目录文件夹
	}
}
```

- `listen` 启动http service时绑定的地址
- `tmp_path` 临时文件的保存路径，一般临时包括：上传图片的原图、待上传到Zurich的文件、待转换的HTML文件，
  这些文件一般会在使用后马上删除，不过也不排除程序问题没有删除的文件。
- `web_root` http service使用的webroot
- `ssh` Zurich sftp的相关登录信息
   - `host` sftp host with port (eg: 127.0.0.1:22)
   - `user` sftp login username
   - `password` sftp login pwd
   - `key` sftp login private key file path
- `deploy_path`  Zurich sftp的发布路径，用于区分不同的运行环境，一般不用更改


## 生成 `swagger` 文档

- 安装 [swagger-go](https://github.com/go-swagger/go-swagger)
- 在项目目录执行
```bash
swagger generate spec -o ./web_root/swagger/swagger.json
```

## Docker

此项目已经打包成docker 镜像

- 签出docker 镜像
```
docker pull mmhk/pgp-sftp-proxy
```
- 环境变量，具体请参考 `config.json` 的说明。
  - HOST，service绑定的服务地址及端口，默认为 `127.0.0.1:3333`
  - ROOT, swagger-ui 存放的本地目录，可以设置空来屏蔽 swagger-ui 的显示， 默认为 `/usr/local/mmhk/pgp-sftp-proxy/web_root`
  - SSH_HOST, SSH远程访问host
  - SSH_USER, SSH远程登录账户
  - SSH_PWD, SSH远程登录密码
  - SSH_KEY, SSH远程登录密匙，当sftp 使用密匙登录的时候使用，是一个本地文件路径。（注意是容器中的路径，应该使用 `-v`参数映射进容器）
  - DEPLOY_PATH_DEV, sftp 远程开发目录文件夹, 默认值：`/Interface_Development_Files/`
  - DEPLOY_PATH_PRODUCTION, sftp 远程产品目录文件夹, 默认值：`/Interface_Production_Files/`
  - DEPLOY_PATH_TESTING, sftp 远程测试目录文件夹, 默认值：`/Interface_UAT_Files/`
- 运行
```
docker run --name mmhk/pgp-sftp-proxy -p 3333:3333 mmhk/mmhk/pgp-sftp-proxy:latest
```
