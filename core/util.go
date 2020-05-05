/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: util.go
 * Date: 2020/5/2 上午1:19
 * Author: sevth
 */

package core

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/chenhg5/collection"
	"github.com/gen2brain/beeep"
	"github.com/sevth-developer/clipboard"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	user2 "os/user"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// 初始化时候，顺便加载log

type Util struct {
	HomePath   string
	WorkDir    string
	ConfigPath string
	LogPath    string
	SavePath   string
	Log        *zap.SugaredLogger
	Except     [][]bool
	Method     string
	ExceptUses []string
}

var util *Util

// 构造工具集合 初始化工具集并且返回一个工具集
func LoadUtil() *Util {
	util = new(Util)
	util.initUtil()
	return util
}

//InitUtil 初始化工具集
func (u *Util) initUtil() {
	u.HomePath = getHomeDir()
	// 设置项目根目录
	u.WorkDir = u.HomePath + "/FCM"
	u.ConfigPath = u.WorkDir + "/config"
	u.LogPath = u.WorkDir + "/log"
	u.SavePath = u.WorkDir + "/save"
	// 设置 log
	u.Log = NewLogger()
	u.ExceptUses = []string{"smms"}
}

// 以下为工具集的方法

// 打开文件
func (u Util) OpenFile(FilePath string) (*os.File, error) {
	return os.OpenFile(FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
}

// 只读方式读取文件
func (u Util) OpenFileByReadOnly(FilePath string) (*os.File, error) {
	return os.Open(filePath)
}

// 检查文件是否存在
func (u Util) IsFileExist(FilePath string) bool {
	if _, err := os.Stat(FilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// 获取文件扩展名
func (u Util) GetFileExt(FilePath string) string {
	var fileExt string
	for i := len(FilePath) - 1; i >= 0; i-- {
		if FilePath[i] == '/' {
			break
		}
		if FilePath[i] == '.' {
			fileExt = FilePath[i:]
			if fileExt == ".gz" || fileExt == ".bz" || fileExt == ".bz2" {
				continue
			}
			break
		}
	}
	return fileExt
}

// 获取文件大小 单位:Byte
func (u Util) GetFileSize(FilePath string) int64 {
	file, _ := u.OpenFile(FilePath)
	fi, _ := file.Stat()
	return fi.Size()
}

// 文件夹判断操作
func (u Util) IsPathExists(path string) error {
	_, err := os.Stat(path)
	return err
}

// 创建文件夹
func (u Util) MakeDIR(path string) error {
	return os.MkdirAll(path, 0755)
}

// 合成带目录的文件路径
func (u Util) MakeFileKey(dir, FilePath string) string {
	k := u.GenerateRandomKey()
	dir = strings.ReplaceAll(dir, "{Y}", time.Now().Format("2006"))
	dir = strings.ReplaceAll(dir, "{y}", time.Now().Format("06"))
	dir = strings.ReplaceAll(dir, "{M}", time.Now().Format("Jan"))
	dir = strings.ReplaceAll(dir, "{m}", time.Now().Format("01"))
	dir = strings.ReplaceAll(dir, "{d}", time.Now().Format("02"))
	dir = strings.ReplaceAll(dir, "{H}", time.Now().Format("15"))
	dir = strings.ReplaceAll(dir, "{h}", time.Now().Format("03"))
	dir = strings.ReplaceAll(dir, "{R}", u.GetArchiveDirName(FilePath))
	if dir[:1] == "/" {
		dir = strings.TrimLeft(dir, "/")
	}
	if dir[len(dir)-1:] != "/" {
		dir = dir + "/"
	}
	key := dir + k + u.GetFileExt(FilePath)
	return key
}

// 生成随机字符串
func (u Util) GenerateRandomKey() string {
	var RLen int8 = 16
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, RLen)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// 获取文件分类对应目录名
func (u Util) GetArchiveDirName(FilePath string) string {
	ext := strings.ToLower(strings.TrimLeft(u.GetFileExt(FilePath), "."))
	pictures := []string{"jpg", "jpeg", "peg", "png", "gif", "tiff", "tif", "webp",
		"svg", "bmp", "ai", "ico", "icns", "ppm", "pgm", "pnm", "pbm", "bgp"}
	music := []string{"mp3", "aac", "wav", "ogg", "flac", "wma", "ac3", "pcm", "aiff",
		"alac", "wpl", "aa", "act", "ape", "m4a", "m4p", "oga", "mogg", "tta"}
	videos := []string{"mkv", "avi", "3gp", "mov", "bik", "wmv", "flv", "swf", "vob",
		"ifo", "mp4", "m4v", "mpg", "asf", "mpeg", "mpv", "qt", "rmvb", "ts"}
	programs := []string{"exe", "apk", "com", "deb", "msi", "dmg", "bin", "vcd",
		"pl", "cgi", "jar", "py", "wsf"}
	documents := []string{"docx", "pdf", "doc", "txt", "rtf", "odt", "tex", "docm",
		"wks", "wps", "ppt", "ods", "pptx", "xlr", "xlt", "xls", "xlsx", "xml", "key",
		"rss", "cer"}
	books := []string{"djvu", "fb2", "fb3", "mobi", "epub", "azw", "lit", "odf", "kfx"}
	archives := []string{"zip", "rar", "7z", "gzip", "gz", "tar", "arj", "rpm", "tar.gz",
		"tar.bz", "tar.bz2"}
	images := []string{"iso", "adf", "cso", "md0", "md1", "md2", "mdf", "cdr"}

	if collection.Collect(pictures).Contains(ext) {
		return "picture"
	}
	if collection.Collect(music).Contains(ext) {
		return "music"
	}
	if collection.Collect(videos).Contains(ext) {
		return "video"
	}
	if collection.Collect(programs).Contains(ext) {
		return "program"
	}
	if collection.Collect(documents).Contains(ext) {
		return "document"
	}
	if collection.Collect(books).Contains(ext) {
		return "book"
	}
	if collection.Collect(archives).Contains(ext) {
		return "archive"
	}
	if collection.Collect(images).Contains(ext) {
		return "images"
	}
	return "other"
}

// 取得不带扩展名的文件名
func (u Util) GetFileNameWithoutExt(FilePath string) string {
	var name string
	ext := u.GetFileExt(FilePath)
	for i := len(FilePath) - 1; i >= 0; i-- {
		if FilePath[i] == '/' {
			name = FilePath[i+1:]
			break
		}
	}
	if len(name) == 0 {
		name = FilePath
	}
	return name[0 : len(name)-len(ext)]
}

// 获取操作系统类型
func (u Util) GetOS() int8 {
	sysType := runtime.GOOS
	if sysType == "darwin" {
		return 1
	}
	if sysType == "linux" {
		return 2
	}
	if sysType == "windows" {
		return 3
	}
	return 0
}

// 获取 MIME
func (u Util) GetFileMimeType(path string) string {
	f, err := u.OpenFile(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	defer func() { _, _ = f.Seek(0, 0) }()
	if err != nil {
		return "plain/text"
	}
	return http.DetectContentType(buffer)
}

// 根据传入的自定义域名，生成返回链接
func (u Util) MakeReturnLink(customDomain, bucketName, Endpoint string) (link string) {
	if customDomain == "" {
		customDomain = "https://" + bucketName + "." + Endpoint
	}
	if customDomain[len(customDomain)-1] == '/' {
		customDomain = customDomain[:len(customDomain)-1]
	}
	link = customDomain + "/" + fileKey
	return
}

// 发送开始上传通知
func (u Util) SendStartUploadNotify(count int) (err error) {
	return beeep.Notify("文件上传中…", strconv.Itoa(count)+"个文件正在上传，请稍等片刻!", "")
}

// 未检测到上传的文件
func (u Util) SendNoFileNotify() (err error) {
	return beeep.Notify("未检测到文件", "未检测到要上传的文件!", "")
}

// 上传完成通知
func (u Util) SendUploadSuccessNotify(save bool) (err error) {
	if !save {
		return beeep.Notify("上传成功", "链接已在剪贴板中，请直接去粘贴", "")
	}
	return beeep.Notify("上传成功", "链接已在剪贴板中，并且已保存到文件，保存路径为:'"+u.SavePath+"'", "")
}

// 上传失败通知
func (u Util) SendUploadFailedNotify() (err error) {
	return beeep.Notify("上传失败", "文件上传失败，请粘贴或去日志看是否有报错信息", "")
}

// 设置剪切板内容
func (u Util) SetClipboard(name, link []string) error {
	var buffer bytes.Buffer
	var temp string
	img := []string{
		"jpg", "jpeg", "png", "gif", "bmp", "ico",
	}
	for i, _ := range name {
		ext := strings.ToLower(strings.TrimLeft(u.GetFileExt(link[i]), "."))
		temp = link[i] + "\n"
		if collection.Collect(img).Contains(ext) {
			temp = "!" + "[" + name[i] + "](" + link[i] + ")\n"
		}
		buffer.WriteString(temp)
	}
	return clipboard.WriteAll(buffer.String())
}

// 获取文件MD5值
func (u Util) GetFileMD5(FilePath string) string {
	f, err := os.Open(FilePath)
	if err != nil {
		return ""
	}
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	return fmt.Sprintf("%x", md5.Sum(b))
}

// 文件基础处理
func (u Util) FileInfo(fp string) (Name, MD5, Mime, path string, Size int64) {
	Name = u.GetFileNameWithoutExt(fp)

	f, err := os.Open(fp)
	if err != nil {
		return
	}
	defer f.Close()
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	defer func() { _, _ = f.Seek(0, 0) }()
	if err != nil {
		Mime = "plain/text"
	}

	MD5 = u.GetFileMD5(fp)
	//fmt.Println(buffer)
	//MD5 = "123456"
	Mime = http.DetectContentType(buffer)
	Size = u.GetFileSize(fp)
	path = fp
	return
}

// 获得用户主目录
func getHomeDir() string {
	// 获取用户主目录路径
	var home string
	user, err := user2.Current()
	if err == nil {
		home = user.HomeDir
	}
	if runtime.GOOS == "windows" && home == "" {
		drive := os.Getenv("HOMEDRIVE")
		homePath := os.Getenv("HOMEPATH")
		home = drive + homePath
		if drive == "" || homePath == "" {
			home = os.Getenv("USERPROFILE")
		}
	}
	if home == "" {
		home = os.Getenv("HOME")
	}
	if home == "" {
		var stdout bytes.Buffer
		cmd := exec.Command("sh", "-c", "eval echo ~$USER")
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil {
			home = strings.TrimSpace(stdout.String())
		}
	}
	if home == "" {
		home = "./"
	}
	return home
}

func (u Util) makeForm(file, field map[string]string) (*bytes.Buffer, string) {
	var buffer = new(bytes.Buffer)
	w := multipart.NewWriter(buffer)
	for keyName, fp := range file {
		fw, err := w.CreateFormFile(keyName, fp)
		if err != nil {
			fmt.Println(err)
		}
		fd, err := u.OpenFileByReadOnly(fp)
		if err != nil {
			fmt.Println(err)
		}
		_, err = io.Copy(fw, fd)
		fd.Close()
	}

	for k, v := range field {
		err := w.WriteField(k, v)
		if err != nil {
			fmt.Println(err)
		}
	}
	c := w.FormDataContentType()

	defer w.Close()
	return buffer, c
}
