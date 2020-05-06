/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: AwsMultipartUpload.go
 * Date: 2020/5/5 下午11:01
 * Author: sevth
 */

package core

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"sync"
)

type AwsMultiPartUpload struct {
	Bucket         string
	FilePath       string
	FileKey        string
	FileMime       string
	FileSize       int64
	PartSize       int64               // 分片大小
	chunkCount     int                 // 分块个数
	Tries          int                 // 单个分片上传尝试次数
	Goroutine      int                 // 线程
	CompletedParts []*s3.CompletedPart // 保存分片数据
	mux            sync.Mutex
}

//func NewAwsMultiPartUpload() *AwsMultiPartUpload {
//	return &AwsMultiPartUpload{
//		Goroutine: 8,
//		Tries:     3,
//		PartSize:  1 * 1024 * 1024,
//	}
//}

func (a *AwsMultiPartUpload) AwsMultipartUpload(svc *s3.S3) error {
	if 16 < a.Goroutine || 0 >= a.Goroutine {
		a.Goroutine = 8 // 不允许设置过大
	}
	if a.Bucket == "" || a.FilePath == "" {
		return errors.New("未设置必须参数")
	}
	if a.Tries == 0 {
		a.Tries = 3
	}
	if a.PartSize == 0 {
		a.PartSize = 4 * 1024 * 1024
	}
	if svc == nil {
		return errors.New("svc 错误")
	}
	a.chunkCount = util.divideCeil(a.FileSize, a.PartSize) // 分块个数
	if len(a.CompletedParts) == 0 {                        //初始化切片，长度为分块个数，后面分片排序需要用到
		a.CompletedParts = make([]*s3.CompletedPart, a.chunkCount)
	}
	return a.awsMultiPartUpload(svc)
}

func (a *AwsMultiPartUpload) awsMultiPartUpload(svc *s3.S3) error {
	file, err := util.OpenFileByReadOnly(a.FilePath)
	if err != nil {
		return err
	}
	buffer := make([]byte, a.FileSize) // 使用一个文件大小长度的byte切片存储上传的文件数据
	buffer, _ = ioutil.ReadAll(file)   // ioutil 读取速度快！
	defer file.Close()

	initData, err := a.initMultipartUpload(svc) // 初始化分块
	if err != nil {
		return err
	}
	errChan := make(chan error, a.Goroutine) // 创建通道
	for i := 0; i != a.Goroutine; i++ {      //通道置空，阻塞八个并发
		errChan <- nil
	}

	wg := &sync.WaitGroup{}
	for i := 0; i != a.chunkCount; i++ {
		//println(i)
		wg.Add(1)
		go func(pos int) { // 第几个块，取偏移量
			defer wg.Done()                     // 完成分片，计数器减一
			start := a.PartSize * int64(pos)    // 起始地址
			offset := a.PartSize * int64(pos+1) // 偏移地址
			if pos == a.chunkCount-1 {          // 最后一个块，偏移地址为文件长度
				offset = a.FileSize
			}
			err := a.uploadPart(svc, initData, buffer[start:offset], pos)
			errChan <- err
		}(i)
		uploadErr := <-errChan // 接收通道值
		if uploadErr != nil {
			err = uploadErr
			break //上传出错，需要取消上传
		}
	}
	wg.Wait() // 等待任务完成

	select { //检查一下是否有剩余的通道未接收，然后检查
	case e := <-errChan:
		if e != nil {
			err = e
		}
	default:
		err = nil
	}
	close(errChan)  // 关闭通道
	if err != nil { // 处理之前的错误
		_, err := a.abortMultipartUpload(svc, initData)
		if err != nil {
			//fmt.Println(resp)
			return err
		}
	}

	_, err = a.completeMultipartUpload(svc, initData, a.CompletedParts) // 分片上传完成
	return err
}

// 初始化分片上传
func (a *AwsMultiPartUpload) initMultipartUpload(svc *s3.S3) (*s3.CreateMultipartUploadOutput, error) {
	return svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{ // 直接返回初始化分片上传的数据
		Bucket:      aws.String(a.Bucket),
		Key:         aws.String(a.FileKey),
		ContentType: aws.String(a.FileMime),
	})
}

// 上传单个分片
func (a *AwsMultiPartUpload) uploadPart(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, fileByte []byte, pos int) (err error) {
	tryNum := 0

	for tryNum < a.Tries {
		uploadResult, err := svc.UploadPart(&s3.UploadPartInput{
			Body:          bytes.NewReader(fileByte),
			Bucket:        resp.Bucket,
			Key:           resp.Key,
			PartNumber:    aws.Int64(int64(pos)),
			UploadId:      resp.UploadId,
			ContentLength: aws.Int64(int64(len(fileByte))),
		})
		if err != nil {
			if tryNum == a.Tries {
				if aerr, ok := err.(awserr.Error); ok {
					return aerr
				}
				return err
			}
			//fmt.Printf("Retrying to upload part #%v\n", pos)
			tryNum++
		} else {
			a.mux.Lock() // 上锁，避免数据出错
			temp := &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(pos)),
			}
			a.CompletedParts[pos] = temp
			//println(a.CompletedParts[pos])
			a.mux.Unlock() //解锁
			return nil     // 不返回的话会无限循环
		}
	}
	return nil
}

// 分片上传完成，请求完成
func (a *AwsMultiPartUpload) completeMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, completedParts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	return svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{ // 完成分片上传，发请求合成文件
		Bucket:          resp.Bucket,
		Key:             resp.Key,
		UploadId:        resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: completedParts},
	})
}

// 分片上传出错，中断上传
func (a *AwsMultiPartUpload) abortMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput) (*s3.AbortMultipartUploadOutput, error) {
	return svc.AbortMultipartUpload(&s3.AbortMultipartUploadInput{ // 分片上传失败调用取消分片上传
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	})
}
