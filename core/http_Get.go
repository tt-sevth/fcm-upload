/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: http_Get.go
 * Date: 2020/5/10 下午5:02
 * Author: sevth
 */

package core

import (
	"errors"
	"net/http"
	"net/url"
)

type getRequest struct {
	client *http.Client
	req    *http.Request
}

type getRequestInputConfig struct {
	Url    string
	Proxy  string
	Client *http.Client
}

func NewGet(c *getRequestInputConfig) (*getRequest, error) {
	var err error
	r := &getRequest{}
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

	if r.client == nil {
		r.client = &http.Client{}
	}
	if c.Client != nil {
		r.client = c.Client
	}

	r.req, err = http.NewRequest("GET", c.Url, nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *getRequest) Send() (*http.Response, error) {
	resp, err := r.client.Do(r.req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
