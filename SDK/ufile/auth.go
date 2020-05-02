package ufile

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"strings"
)

/* Auth 构造签名的工具，使用本 SDK 你不需要关心 Auth 这一整个模块，以及它暴露出来的签名算法。
 * 如果你希望自己封装 API 可以使用这里面的暴露出来的接口来填充 http authorization。
**/

type Auth struct {
	publicKey  string
	privateKey string
}

func NewAuth(publicKey, privateKey string) Auth {
	return Auth{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

/* Authorization 构造一个主要用于上传文件的签名，返回 HMAC-Sh1 的签名字符串，可以直接填充到 HTTP authorization header 里面。
 * key 是传到 ufile 所使用的文件名，bucekt 是文件上传后存放的 bucket。
 * method 就是你当前这个 HTTP 请求的 Method。
 * header 就是你当前这个 HTTP 的 header。
**/

func (A Auth) Authorization(method, bucket, key string, header http.Header) string {
	var sigData string
	method = strings.ToUpper(method)

	md5 := header.Get("Content-MD5")
	contentType := header.Get("Content-Type")
	date := header.Get("Date")

	sigData = method + "\n" + md5 + "\n" + contentType + "\n" + date + "\n"
	resource := "/" + bucket + "/" + key
	sigData += resource

	signature := A.signature(sigData)

	return "UCloud " + A.publicKey + ":" + signature
}

/* AuthorizationPrivateURL 构造私有空间文件下载链接的签名，其中 expires 是当前的时间加上一个过期的时间，再转为字符串。格式是 unix time second.
 * 其他的参数含义和 Authoriazation 函数一样。
 * 有时我们需要把签名后的 URL 直接拿来用，header 参数可以直接构造一个空的 http.Header{} 传入即可。
**/

func (A Auth) AuthorizationPrivateURL(method, bucket, key, expires string, header http.Header) (string, string) {
	var sigData string
	method = strings.ToUpper(method)
	md5 := header.Get("Content-MD5")
	contentType := header.Get("Content-Type")

	sigData = method + "\n" + md5 + "\n" + contentType + "\n" + expires + "\n"
	resource := "/" + bucket + "/" + key
	sigData += resource

	signature := A.signature(sigData)

	return signature, A.publicKey
}

func (A Auth) signature(data string) string {
	mac := hmac.New(sha1.New, []byte(A.privateKey))
	mac.Write([]byte(data))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
