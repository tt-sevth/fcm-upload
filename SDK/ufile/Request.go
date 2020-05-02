package ufile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

/* Request SDK 主要的 request 模块。本 SDK 遵从以下原则：
 *
 * 1.接口尽可能简洁，隐藏所有复杂实现。
 *
 * 2.本 SDK 主要的作用是封装 HTTP 请求，不做过多其他的封装（如 HTTP Body 序列化，详细的错误检查）。
 *
 * 3.只简单封装 HTTP 请求所需要的参数，给接口使用者提供所有原生的 HTTP response header,body,status code 返回，以便排错。
 *
 * 4.远端请求返回值统一返回一个 error，如果为 nil 表示无错。LastResponseStatus，LastResponseHeader，LastResponseBody 可以查看具体的 HTTP 返回信息（）。如果你想少敲几行代码可以直接调用 DumpResponse(true) 查看详细返回。
**/

type Request struct {
	Auth				Auth
	BucketName			string
	Host				string
	Client				*http.Client
	Context 			context.Context
	baseURL				*url.URL
	RequestHeader		http.Header
	err 				error

	LastResponseStatus 	int
	LastResponseHeader 	http.Header
	LastResponseBody	[]byte
	VerifyUploadMD5 	bool
	LastResponse 		*http.Response
}

/* NewFileRequest 创建一个用于管理文件的 request，管理文件的 url 与 管理 bucket 接口不一样，
 * 请将 bucket 和文件管理所需要的分开，NewUBucketRequest 是用来管理 bucket 的。
 * Request 创建后的 instance 不是线程安全的，如果你需要做并发的操作，请创建多个 UFileRequest。
 * config 参数里面包含了公私钥，以及其他必填的参数。详情见 config 相关文档。
 * client 这里你可以传空，会使用默认的 http.Client。如果你需要设置超时以及一些其他相关的网络配置选项请传入一个自定义的 client。
**/

func NewFileRequest(config *Config, client *http.Client) (*Request, error) {
	if config.FileHost == "" || config.BucketName == "" {
		return nil, errors.New("配置文件必须填写 bucket 名字与管理 host 域名")
	}
	request := newRequest(config.PublicKey, config.PrivateKey, config.BucketName, config.FileHost, client)
	request.VerifyUploadMD5 = config.VerifyUploadMD5

	if request.baseURL.Scheme == "" {
		request.baseURL.Host = request.BucketName + "." + request.Host
		request.baseURL.Scheme = "http"
	}
	return request, nil
}

func newRequest(publicKey, privateKey, bucket, host string, client *http.Client) *Request {
	publicKey = strings.TrimSpace(publicKey)
	privateKey = strings.TrimSpace(privateKey)
	bucket = strings.TrimSpace(bucket)
	host = strings.TrimSpace(host)

	request := new(Request)
	request.Auth = NewAuth(publicKey, privateKey)
	request.BucketName = bucket
	request.Host = host
	request.baseURL, request.err = url.Parse(request.Host)
	if request.err != nil {
		panic(request.err)
	}
	request.baseURL.Path = "/"

	if client == nil {
		client = new(http.Client)
	}
	request.Client = client
	request.Context = context.TODO()
	return request
}

func (u *Request) DumpResponse(isDumpBody bool) []byte {
	var b bytes.Buffer
	if u.LastResponse == nil {
		return nil
	}
	b.WriteString(fmt.Sprintf("%s %d\n", u.LastResponse.Proto, u.LastResponseStatus))
	for k, vs := range u.LastResponseHeader {
		str := k + ": "
		for i, v := range vs {
			if i != 0 {
				str += "; " + v
			} else {
				str += v
			}
		}
		b.WriteString(str)
	}
	if isDumpBody {
		b.Write(u.LastResponseBody)
	}
	return b.Bytes()
}

func (u *Request) responseParse(response *http.Response) error {
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	u.LastResponseStatus = response.StatusCode
	u.LastResponseHeader = response.Header
	u.LastResponseBody = responseBody
	u.LastResponse = response
	return nil
}

func (u *Request) requestWithResponse(r *http.Request) (response *http.Response, err error) {
	r.Header.Set("User-Agent", "UFileGoSDK/2.01")

	response, err = u.Client.Do(r.WithContext(u.Context))
	// If we got an error, and the context has been canceled,
	// the context's error is probably more useful.
	if err !=nil {
		select {
		case <-u.Context.Done():
			err = u.Context.Err()
		default:
		}
		return
	}
	return
}

func (u *Request) request(r *http.Request) error {
	response, err := u.requestWithResponse(r)
	if err != nil {
		return err
	}
	err = u.responseParse(response)
	if err != nil {
		return err
	}
	if !VerifyHTTPCode(response.StatusCode) {
		return fmt.Errorf("Remote response code is %d - %s not 2xx call DumpResponse(true) show details",
			response.StatusCode, http.StatusText(response.StatusCode))
	}
	return nil
}
