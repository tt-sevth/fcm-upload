/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: http_Post.go
 * Date: 2020/5/6 下午1:03
 * Author: sevth
 */

package core

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

type request struct {
	client *http.Client
	req    *http.Request
	resp   http.Response
}

type RequestInputConfig struct {
	Url    string
	Proxy  string
	Client *http.Client
	Body   *RequestBodyField
}
type RequestBodyField struct {
	file  map[string]string
	field map[string]string
}

func NewPost(c *RequestInputConfig) (*request, error) {
	var err error
	r := &request{}
	// 检测url情况
	if c.Url == "" {
		return nil, errors.New("url is not set")
	}

	if c.Proxy != "" {
		r.client = &http.Client{Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(c.Proxy)
			},
		}}
	}
	// 自定义client权重更高，设置了client的话，再设置proxy无效
	r.client = &http.Client{}
	if c.Client != nil {
		r.client = c.Client
	}

	body := &bytes.Buffer{}
	bw := multipart.NewWriter(body)

	for keyName, fp := range c.Body.file {
		fw, err := bw.CreateFormFile(keyName, fp)
		if err != nil {
			return nil, err
			//fmt.Println(err)
		}
		fd, err := os.Open(fp)
		if err != nil {
			return nil, err
			//fmt.Println(err)
		}
		_, err = io.Copy(fw, fd)
		fd.Close()
	}

	for k, v := range c.Body.field {
		err := bw.WriteField(k, v)
		if err != nil {
			return nil, err
			//fmt.Println(err)
		}
	}
	bw.Close() // 写完数据直接关闭，不然数据长度校验会出错

	r.req, err = http.NewRequest("POST", c.Url, body)
	if err != nil {
		return nil, err
	}
	r.req.Header.Set("Content-Type", bw.FormDataContentType())
	return r, nil
}

func (r *request) SetHeader(name, value string) {
	r.req.Header.Set(name, value)
}

func (r *request) AddHeader(name, value string) {
	r.req.Header.Add(name, value)
}

func (r *request) Send() (*http.Response, error) {
	resp, err := r.client.Do(r.req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
