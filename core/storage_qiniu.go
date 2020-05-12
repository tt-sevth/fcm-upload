/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: qiniu.go
 * Date: 2020/5/4 下午5:31
 * Author: sevth
 */

package core

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"strings"
)

type Qiniu struct {
	Name         string `json:"name"`
	AK           string `json:"ak"`
	SK           string `json:"sk"`
	BucketName   string `json:"bucket_name"`
	Endpoint     string `json:"endpoint"`
	CustomDomain string `json:"custom_domain"`
}

func (q Qiniu) upload(info *fileInfo) (link string) {

	var err error
	var region string
	if q.Endpoint != "" { //获取地域名字
		region = strings.Split(q.Endpoint, ".")[0]
		region = region[3:]
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(q.Endpoint),
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(q.AK, q.SK, ""),
	}))

	svc := s3.New(sess)

	// 普通上传
	if info.fileSize <= maxFileSize {
		fd, _ := util.OpenFileByReadOnly(info.filePath)
		defer fd.Close()
		_, err = svc.PutObject(&s3.PutObjectInput{
			Body:   aws.ReadSeekCloser(fd),
			Bucket: aws.String(q.BucketName),
			ContentMD5: aws.String(info.md5Header),		//添加MD5校验会失败，暂时不知道原因
			Key: aws.String(info.fileKey),
		})
	} else {
		// 分片上传
		util.Log.Info("七牛云使用分片上传文件：", info.fileName)
		upload := &AwsMultiPartUpload{
			Bucket:   q.BucketName,
			FilePath: info.filePath,
			FileSize: info.fileSize,
			FileKey:  info.fileKey,
			FileMime: info.fileMime,
			PartSize: partSize,
		}
		err = upload.AwsMultipartUpload(svc)
	}

	if err != nil {
		util.Log.Error("JD By AWS SDK throw err ", err)
		return
	}

	return util.MakeReturnLink(q.CustomDomain, q.BucketName, q.Endpoint, info.fileKey)
}

func (q Qiniu) delete(info *fileInfo) bool {
	var err error
	var region string
	if q.Endpoint != "" {
		region = strings.Split(q.Endpoint, ".")[0]
		region = region[3:]
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(q.Endpoint),
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(q.AK, q.SK, ""),
	}))

	svc := s3.New(sess)
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(q.BucketName),
		Key:    aws.String(info.fileKey),
	})
	if err != nil {
		return false
	}
	return true
}
