# FCM

> 暂时没想好取什么名字，姑且这么叫吧 :smirk:

## 简介

这是一个用`golang`编写的多云上传工具，灵感来自于 [`PicUploader`](https://github.com/sevth-developer/PicUploader) , 感谢[`@xiebruce`](https://github.com/xiebruce)。

目前的样子是一个命令行的工具，以后“可能”（挖坑）会有界面吧。同时也是作为自己学习过程的一个见证。目前支持的服务商(对象存储或者图床服务商)如下:

|   服务商    | 文件类型 |
| :---------: | :------: |
|   阿里云    |   任意   |
|   腾讯云    |   任意   |
|   百度云    |   任意   |
|   京东云    |   任意   |
|   ucloud    |   任意   |
|    七牛     |   任意   |
|   又拍云    |   任意   |
|    sm.ms    |   图片   |
| 码云(gitee) |   图片   |

## 特点

- 数据库去重复支持，上传的时候如果数据库已存在记录，会自动跳过上传，并返回数据库中存在的链接。
- 高并发上传大文件支持，大于`128M`的文件都使用分片上传，充分利用你的宽带。
- 跨平台支持，除了配置文件，无需额外环境设置。
- 轻量化，不驻后台。
- 支持删除文件同步到云端(仅指数据库记录与云端，本地文件不作处理)

## 使用说明

```bash
 ./fcm  <option> [args]
```

|    指令    |  参数   |                             说明                             |
| :--------: | :-----: | :----------------------------------------------------------: |
| -i  --init | config  |          初始化配置文件,不出意外会自动打开配置文件           |
|            |   All   |    初始化所有，包括目录，会在用户主目录下建立`FCM`文件夹     |
| -u   --use | console | 使用终端模式上传文件，这个模式在上传完成后打印出所有的文件链接。例如：`-u console ./test.txt` |
|            | system  | 以系统调用方式上传文件，主要用来支持右键上传，上传完成链接后会自动写入剪切板，如果上传多个云空间，只会返回每个文件的第一条链接 |
|            | typora  |                  作为typora的自定义上传插件                  |
|  -d  --db  |  Dump   | 从数据库中导出所有的文件链接，文件保存在`FCM`文件夹的`save`文件夹下 |
|            |  query  |                      查询单个文件的记录                      |
|    -del    |         | 删除某个文件记录，同步删除云端的文件。直接接文件路径 -del /path/to/file |

第一次使用时，执行 `./fcm -i all` 初始化，然后填写配置文件，如果系统没有自动打开文件，请在当前用户主目录下的`/FCM/config/config.json`打开手动编辑。

配置文件大致结构如下：

```
{
  "name": "FCM 配置文件",
  "storage_types": {
    "aliyun": {
      "name": "Aliyun oss SDK",	//无关紧要的参数
      "access_key_id": "LT****************KS",	//云平台获取的key
      "access_key_secret": "Tu*****************e0",	// 云平台获取的secret
      "bucket_name": "bucket",	// 对象存储名称
      "endpoint": "oss-cn-shenzhen.aliyuncs.com",	//地域
      "custom_domain": ""	//自定义返回链接，不填默认使用 bucket_name.enpoint 拼接
    },
    ...		// 省略大致相同的结构体
    "smms": {
      "name": "smms",
      "access_token": "EVYkI2DGsBGcWnt8LK4AtGoGag3qcyQY",	//smms的token
      "proxy": ""
    },
    "gitee": {
      "name": "gitee",
      "owner": "sevth",	// 所有者，就是链接地址的那个名字
      "repo": "image",	// 仓库名字
      "access_token": "0*************2"	//token
    }
  },
  "dir说明": "存放的文件目录 {R} 根据文件后缀判断文件类型，使用对应的路径，时间格式 {Y} 2020 {y} 20 {M} Apr {m} 04 {d} 01",
  "directory": "test/{Y}",// 类似的 {R}/{Y}/{m}/{d} 会自动替换成类似 image/2020/5/1 的形式，{R}是根据文件类型判断的。
  "primary_domain": "",	// 主域名，除非你用反代，具体看 picUpload 的说明
  "uses": [		//上传到哪些(或者一个)服务商
    "gitee",
    "aliyun"		
  ],
  "dsn": {	// 数据库配置，一般除非必要，保存默认就好。
    "uses": "sqlite3",	// 还支持 mysql,mssql,postgres
    "protocol": "",	//类似于 127.0.0.1:3306 / localhost
    "username": "",	//数据库用户名
    "password": "",	//密码
    "dbname": "",	//使用的数据库
    "dsn_link": "",	//dsn 链接  权重高，填写即使用dsn链接连接数据库
    "debug": false
  }
}
```

### 直接使用

- 在终端中输入

  ```bash
  ./fcm -u console /path/to/you/file1 /path/to/you/file2
  ```

- 在系统中使用

  设置右键方式参考[picUpload](https://github.com/sevth-developer/PicUploader) ,指令换成类似如下 

  ```
  /YouHomeDir/FCM/fcm -u system "$@" | pbcopy
  ```

- 在typora中使用

  打开`typora`设置，在图像选项卡，上传服务设定选择 `Custom Command` 自定义命令如下:

  ```
  /YouHomeDir/FCM/fcm -u typora
  ```

  ![image-20200506210347096](https://img.sevth.com/test/2020/LdvdVVVWHlMbSFIC.png)
  
- 删除文件

```bash
./fcm -del /path/to/file
```



## 下载

根据对应系统下载对应的版本 [releases](https://github.com/sevth-developer/FCM/releases)

## TODO

1. 继续修bug，完善项目。

1. 支持数据库与云空间绑定，删除数据库某条数据时，从云空间也删除。
2. 支持更多的数据库操作。

## 支持

直接提交 [issues](https://github.com/sevth-developer/FCM/issues)

到我的博客[留言](https://sevth.com/message/)
