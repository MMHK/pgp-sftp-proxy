# pgp-sftp-proxy [for DahSing]
sftp http proxy with PGP encrypt


具体实现的功能如下:

- PGP加密文件
- 上传文件并PGP加密， 到`DahSing`的 `SFTP`
- 自带http server，使用http rest API操作
- API文档请编译后执行 `http://127.0.0.1:3334/swagger/index.html`


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

```json
{
    "listen": "127.0.0.1:3334", //服务绑定的host
    "tmp_path": "./web_root/temp", //文件缓存目录
	"web_root" : "./web_root", //API 文档说明目录
	"ssh" : {
		"host" : "", //ssh 远程登录host
		"user" : "", //ssh 远程登录账户
		"password" : "", //ssh 远程登录密码
		"key" : "" //ssh 远程登录密匙
	},
	"time-range": [
	  {
	    "begin": "00:00",
	    "end": "18:00"
	  },
	  {
	    "begin": "22:00",
	    "end": "00:00"
	  }
	]
}
```

- `listen` 启动http service时绑定的地址
- `tmp_path` 临时文件的保存路径，一般临时包括：上传图片的原图、待上传到DahSing的文件、待转换的HTML文件，
  这些文件一般会在使用后马上删除，不过也不排除程序问题没有删除的文件。
- `web_root` http service使用的webroot
- `ssh` DahSing sftp的相关登录信息
   - `host` sftp host with port (eg: 127.0.0.1:22)
   - `user` sftp login username
   - `password` sftp login pwd
   - `key` sftp login private key file path
- `time-range`  DahSing 限定可以上传的时间段，用于避开DahSing 处理 sftp 的时间段 


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
docker pull mmhk/pgp-sftp-proxy:dahsing
```
- 环境变量，具体请参考 `config.json` 的说明。
  - HOST，service绑定的服务地址及端口，默认为 `127.0.0.1:3334`
  - ROOT, swagger-ui 存放的本地目录，可以设置空来屏蔽 swagger-ui 的显示， 默认为 `/usr/local/mmhk/pgp-sftp-proxy/web_root`
  - SSH_HOST, SSH远程访问host
  - SSH_USER, SSH远程登录账户
  - SSH_PWD, SSH远程登录密码
  - SSH_KEY, SSH远程登录密匙，当sftp 使用密匙登录的时候使用，是一个本地文件路径。（注意是容器中的路径，应该使用 `-v`参数映射进容器）
- 运行
```
docker run --name mmhk/pgp-sftp-proxy:dahsing -p 3334:3334 mmhk/mmhk/pgp-sftp-proxy:latest
```
