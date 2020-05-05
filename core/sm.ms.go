/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: sm.ms.go
 * Date: 2020/5/5 上午1:37
 * Author: sevth
 */

package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Smms struct {
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
	Proxy       string `json:"proxy"`
}

func smms() (link string) {
	if fileSize > 5<<20 {
		return
	}
	body, contentType := util.makeForm(map[string]string{"smfile": filePath}, map[string]string{"format": "json"})
	token := config.StorageTypes.Smms.AccessToken
	if token == "" {
		token = "EVYkI2DGsBGcWnt8LK4AtGoGag3qcyQY"
	}

	req, err := http.NewRequest("POST", "https://sm.ms/api/v2/upload", body)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	if config.StorageTypes.Smms.Proxy != "" {
		client = &http.Client{Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(config.StorageTypes.Smms.Proxy)
			},
		}}
	}
	//resp, err := http.DefaultClient.Do(req)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
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
