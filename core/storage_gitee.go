/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: storage_gitee.go
 * Date: 2020/5/6 下午12:49
 * Author: sevth
 */

package core

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

type Gitee struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	AccessToken string `json:"access_token"`
}

func (g Gitee) upload(info *fileInfo) (link string) {
	if info.fileSize > 5<<21 {
		return ""
	}
	url := "https://gitee.com/api/v5/repos/" + g.Owner + "/" + g.Repo + "/contents/" + info.fileKey

	fd, err := os.Open(info.filePath)
	if err != nil {
		util.Log.Error("gitee throw err ", err)
		return
	}
	defer fd.Close()
	bytes, err := ioutil.ReadAll(fd)
	if err != nil {
		util.Log.Error("gitee throw err ", err)
		return
	}

	post, err := NewPost(&PostRequestInputConfig{
		Url: url,
		Body: &PostRequestBodyField{
			file: nil,
			field: map[string]string{
				"access_token": g.AccessToken,
				"content":      util.Base64Content(bytes),
				"message":      "post image：" + info.fileKey,
			},
		},
	})
	if err != nil {
		util.Log.Error("gitee throw err ", err)
		return
	}
	resp, err := post.Send()
	if err != nil {
		util.Log.Error("gitee throw err ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		util.Log.Error("gitee throw err, resp statusCode: ", err)
		return
	}

	return "https://gitee.com/" + g.Owner + "/" + g.Repo + "/raw/master/" + info.fileKey
}

func (g Gitee) delete(info *fileInfo) bool {
	type shaResp struct {
		Sha string `json:"sha"`
	}
	url := "https://gitee.com/api/v5/repos/" + g.Owner + "/" + g.Repo + "/contents/" + info.fileKey + "?access_token=" + g.AccessToken
	resp, err := http.Get(url)

	if err != nil {
		util.Log.Error("gitee throw err ", err)
		return false
	}

	defer resp.Body.Close()
	shaR := &shaResp{}
	respBody, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(respBody, shaR)

	if shaR.Sha == "" {
		return false
	}

	deleteUrl := url + "&sha=" + shaR.Sha + "&message=delete+image：" + info.fileKey
	req, err := http.NewRequest(http.MethodDelete, deleteUrl, nil)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		util.Log.Error("gitee throw err ", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		util.Log.Error("gitee throw err, statusCode is ", resp.StatusCode)
		return false
	}
	return true
}
