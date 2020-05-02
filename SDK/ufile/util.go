package ufile

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	blkSiZE = 2 << 21
)

//VerifyHTTPCode 检查 HTTP 的返回值是否为 2XX，如果不是就返回 false。
func VerifyHTTPCode(code int) bool {
	if code < http.StatusOK || code > http.StatusIMUsed {
		return false
	}
	return true
}

//GetFileMimeType 获取文件的 mime type 值，接收文件路径作为参数。如果检测不到，则返回空。

func GetFileMimeType(path string) string {
	f, err := openFile(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	return getMimeType(f)
}

func getMimeType(f *os.File) string {
	buffer := make([]byte, 512)
	_, err := f.Read(buffer)
	defer func() { f.Seek(0, 0) }()
	if err != nil {
		return "plain/text"
	}
	return http.DetectContentType(buffer)
}

func openFile(path string) (*os.File, error) {
	return os.Open(path)
}

//getFileSize get opened file size

func getFileSize(f *os.File) int64 {
	file, err := f.Stat()
	if err != nil {
		panic(err.Error())
	}
	return file.Size()
}

// 获取文件的 etag 值

func GetFileEtag(path string) string {
	f, err := openFile(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	return calculateEtag(f)
}

// 计算文件的 etag 值

func calculateEtag(f *os.File) string {
	fsize := getFileSize(f)
	blkcnt := uint32(fsize / blkSiZE)
	if fsize%blkSiZE !=0 {
		blkcnt++
	}

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, blkcnt)

	h := sha1.New()
	buf := make([]byte, 0, 24)
	buf = append(buf, bs...)
	if fsize <= blkSiZE {
		io.Copy(h, f)
	} else {
		var i uint32
		for i = 0; i < blkcnt; i++ {
			shaBlk := sha1.New()
			io.Copy(shaBlk, io.LimitReader(f, blkSiZE))
			io.Copy(h, bytes.NewReader(shaBlk.Sum(nil)))
		}
	}
	buf = h.Sum(buf)
	etag := base64.URLEncoding.EncodeToString(buf)
	return etag
}

func structPrettyStr(data interface{}) string {
	bytes, err := json.MarshalIndent(data, "", " ")
	if err !=nil {
		return ""
	}
	return fmt.Sprintf("%s\n", bytes)
}