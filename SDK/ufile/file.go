package ufile

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	fourMegabyte = 1 << 21 //4M
)

//FileDataSet  用于 FileListResponse 里面的 DataSet 字段。
type FileDataSet struct {
	BucketName    string `json:"BucketName,omitempty"`
	FileName      string `json:"FileName,omitempty"`
	Hash          string `json:"Hash,omitempty"`
	MimeType      string `json:"MimeType,omitempty"`
	FirstObject   string `json:"first_object,omitempty"`
	Size          int    `json:"Size,omitempty"`
	CreateTime    int    `json:"CreateTime,omitempty"`
	ModifyTime    int    `json:"ModifyTime,omitempty"`
	StorageClass  string `json:"StorageClass,omitempty"`
	RestoreStatus string `json:"RestoreStatus,omitempty"`
}

//FileListResponse 用 PrefixFileList 接口返回的 list 数据。
type FileListResponse struct {
	BucketName string        `json:"BucketName,omitempty"`
	BucketID   string        `json:"BucketId,omitempty"`
	NextMarker string        `json:"NextMarker,omitempty"`
	DataSet    []FileDataSet `json:"DataSet,omitempty"`
}

func (f FileListResponse) String() string {
	return structPrettyStr(f)
}

// 文件秒传
func (u *Request) UploadHit(filePath, keyName string) (err error) {
	file, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	fsize := getFileSize(file)
	etag := calculateEtag(file)

	query := &url.Values{}
	query.Add("Hash", etag)
	query.Add("FileName", keyName)
	query.Add("FileSize", strconv.FormatInt(fsize, 10))
	requestURL := u.genFileURL("uploadhit") + "?" + query.Encode()

	r, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("POST", u.BucketName, keyName, r.Header)
	r.Header.Add("authorization", authorization)

	return u.request(r)
}

func (u *Request) genFileURL(keyName string) string {
	return u.baseURL.String() + keyName
}

/*PutFile 把文件直接放到 HTTP Body 里面上传，相对 PostFile 接口，这个要更简单，速度会更快（因为不用包装 form）。
mimeType 如果为空的，会调用 net/http 里面的 DetectContentType 进行检测。
keyName 表示传到 ufile 的文件名。
小于 100M 的文件推荐使用本接口上传。*/

func (u *Request) PutFile(filePath, keyName, mimeType string) error {
	requestURL := u.genFileURL(keyName)
	file, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	r, err := http.NewRequest("PUT", requestURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	if mimeType == ""{
		mimeType = GetFileMimeType(filePath)
	}
	r.Header.Add("Content-Type", mimeType)
	for k, v := range u.RequestHeader {
		for i := 0; i < len(v); i++ {
			r.Header.Add(k, v[i])
		}
	}

	if u.VerifyUploadMD5 {
		md5Str := fmt.Sprintf("%x", md5.Sum(b))
		r.Header.Add("Content-MD5", md5Str)
	}

	authorization := u.Auth.Authorization("PUT", u.BucketName, keyName, r.Header)
	r.Header.Add("authorization", authorization)
	fileSize := getFileSize(file)
	r.Header.Add("Content-Length", strconv.FormatInt(fileSize, 10))

	return u.request(r)
}

//DeleteFile 删除一个文件，如果删除成功 statuscode 会返回 204，否则会返回 404 表示文件不存在。
//keyName 表示传到 ufile 的文件名。
func (u *Request) DeleteFile(keyName string) error {
	reqURL := u.genFileURL(keyName)
	r, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("DELETE", u.BucketName, keyName, r.Header)
	r.Header.Add("authorization", authorization)
	return u.request(r)
}

//HeadFile 获取一个文件的基本信息，返回的信息全在 header 里面。包含 mimeType, content-length（文件大小）, etag, Last-Modified:。
//keyName 表示传到 ufile 的文件名。
func (u *Request) HeadFile(keyName string) error {
	reqURL := u.genFileURL(keyName)
	r, err := http.NewRequest("HEAD", reqURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("HEAD", u.BucketName, keyName, r.Header)
	r.Header.Add("authorization", authorization)
	return u.request(r)
}

//PrefixFileList 获取文件列表。
//prefix 表示匹配文件前缀。
//marker 标志字符串
//limit 列表数量限制，传 0 会默认设置为 20.
func (u *Request) PrefixFileList(prefix, marker string, limit int) (list FileListResponse, err error) {
	query := &url.Values{}
	query.Add("prefix", prefix)
	query.Add("marker", marker)
	if limit == 0 {
		limit = 20
	}
	query.Add("limit", strconv.Itoa(limit))
	reqURL := u.genFileURL("") + "?list&" + query.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	authorization := u.Auth.Authorization("GET", u.BucketName, "", req.Header)
	req.Header.Add("authorization", authorization)

	err = u.request(req)
	if err != nil {
		return
	}
	err = json.Unmarshal(u.LastResponseBody, &list)
	return
}

//GetPublicURL 获取公有空间的文件下载 URL
//keyName 表示传到 ufile 的文件名。
func (u *Request) GetPublicURL(keyName string) string {
	return u.genFileURL(keyName)
}

//GetPrivateURL 获取私有空间的文件下载 URL。
//keyName 表示传到 ufile 的文件名。
//expiresDuation 表示下载链接的过期时间，从现在算起，24 * time.Hour 表示过期时间为一天。
func (u *Request) GetPrivateURL(keyName string, expiresDuation time.Duration) string {
	t := time.Now()
	t = t.Add(expiresDuation)
	expires := strconv.FormatInt(t.Unix(), 10)
	signature, publicKey := u.Auth.AuthorizationPrivateURL("GET", u.BucketName, keyName, expires, http.Header{})
	query := url.Values{}
	query.Add("UCloudPublicKey", publicKey)
	query.Add("Signature", signature)
	query.Add("Expires", expires)
	reqURL := u.genFileURL(keyName)
	return reqURL + "?" + query.Encode()
}

//CompareFileEtag 检查远程文件的 etag 和本地文件的 etag 是否一致
func (u *Request) CompareFileEtag(remoteKeyName, localFilePath string) bool {
	err := u.HeadFile(remoteKeyName)
	if err != nil {
		return false
	}
	remoteEtag := strings.Trim(u.LastResponseHeader.Get("Etag"), "\"")
	localEtag := GetFileEtag(localFilePath)
	return remoteEtag == localEtag
}

//Rename 重命名指定文件
//keyName 需要被重命名的源文件
//newKeyName 修改后的新文件名
//force 如果已存在同名文件，值为"true"则覆盖，否则会操作失败
func (u *Request) Rename(keyName, newKeyName, force string) (err error) {

	query := url.Values{}
	query.Add("newFileName", newKeyName)
	query.Add("force", force)
	reqURL := u.genFileURL(keyName) + "?" + query.Encode()

	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("PUT", u.BucketName, keyName, req.Header)
	req.Header.Add("authorization", authorization)
	return u.request(req)
}

//Copy 从同组织下的源Bucket中拷贝指定文件到目的Bucket中，并以新文件名命名
//dstkeyName 拷贝到目的Bucket后的新文件名
//srcBucketName 待拷贝文件所在的源Bucket名称
//srcKeyName 待拷贝文件名称
func (u *Request) Copy(dstkeyName, srcBucketName, srcKeyName string) (err error) {

	reqURL := u.genFileURL(dstkeyName)

	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Ufile-Copy-Source", "/" + srcBucketName + "/" + srcKeyName)

	authorization := u.Auth.Authorization("PUT", u.BucketName, dstkeyName, req.Header)
	req.Header.Add("authorization", authorization)
	return u.request(req)
}
