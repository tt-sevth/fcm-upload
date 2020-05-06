/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: storage_gitee.go
 * Date: 2020/5/6 下午12:49
 * Author: sevth
 */

package core

import (
	"io/ioutil"
	"os"
)

type Gitee struct {
	Name        string	`json:"name"`
	Owner       string	`json:"owner"`
	Repo        string	`json:"repo"`
	AccessToken string	`json:"access_token"`
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

	post, err := NewPost(&RequestInputConfig{
		Url: url,
		Body: &RequestBodyField{
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
