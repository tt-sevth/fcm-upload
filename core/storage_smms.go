/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: sm.ms.go
 * Date: 2020/5/5 上午1:37
 * Author: sevth
 */

package core

import (
	"encoding/json"
	"io/ioutil"
)

type Smms struct {
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
	Proxy       string `json:"proxy"`
}

func (s Smms) upload(info *fileInfo) (link string) {
	util.Log.Info("smms upload")
	if info.fileSize > 5<<20 {
		return
	}
	token := s.AccessToken

	post, err := NewPost(&PostRequestInputConfig{
		Url:    "https://sm.ms/api/v2/upload",
		Proxy:  config.StorageTypes.Smms.Proxy,
		Client: nil,
		Body: &PostRequestBodyField{
			file:  map[string]string{"smfile": info.filePath},
			field: map[string]string{"format": "json"},
		},
	})
	if err != nil {
		util.Log.Error("smms throw err ", err)
		return
	}
	post.SetHeader("Authorization", token)

	resp, err := post.Send()

	if err != nil {
		util.Log.Error("smms throw err ", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		util.Log.Error("smms resp statusCode ", resp.StatusCode)
		return
	}

	type Result struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Images  string `json:"images"`
		Data    struct {
			Url string `json:"url"`
		} `json:"data"`
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	result := &Result{}
	_ = json.Unmarshal(respBody, result)

	//fmt.Println(string(respBody))
	if !result.Success {
		if result.Images != "" {
			return result.Images
		}
		util.Log.Error("throw err ", result.Message)
		return
	}
	return result.Data.Url
}
