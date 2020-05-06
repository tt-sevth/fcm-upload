/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: JD.go
 * Date: 2020/5/5 下午12:52
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

type JD struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	CustomDomain    string `json:"custom_domain"`
}

func (jd JD)upload(info *fileInfo) (link string) {
	var err error
	var region string
	if jd.Endpoint != "" {
		region = strings.Split(jd.Endpoint, ".")[1]
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(jd.Endpoint),
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(jd.AccessKeyId, jd.AccessKeySecret, ""),
	}))

	svc := s3.New(sess)

	// 普通上传
	if info.fileSize <= maxFileSize {
		fd, _ := util.OpenFileByReadOnly(info.filePath)
		defer fd.Close()
		_, err = svc.PutObject(&s3.PutObjectInput{
			Body:       aws.ReadSeekCloser(fd),
			Bucket:     aws.String(jd.BucketName),
			ContentMD5: aws.String(info.fileMD5),
			Key:        aws.String(info.fileKey),
		})
	} else {
		// 分片上传
		util.Log.Info("京东云使用分片上传文件：", info.fileName)
		//upload := NewAwsMultiPartUpload()
		upload := &AwsMultiPartUpload{
			Bucket:   jd.BucketName,
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
	return util.MakeReturnLink(jd.CustomDomain, jd.BucketName, jd.Endpoint, info.fileKey)
}
