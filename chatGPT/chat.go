package chatGPT

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const HOST = "https://api.openai.com/v1"

type ApiKey string

type Uri string

type Config struct {
	Uri string
	ApiKey string
}

type Message struct {
	Role string `json:"role"` // 消息作者 One of system, user, or assistant.
	Content string `json:"content"` // 文本内容
	Name string `json:"name"`
}

type Parameters struct {
	Model string `json:"model"`
	Stream bool `json:"stream"` // 是否流的形式返回
	Temperature float64 `json:"temperature"` // 回答性格 介于 0 和 2 之间。较高的值（如 0.8）将使输出更加随机，而较低的值（如 0.2）将使输出更加集中和确定
	Messages []Message `json:"messages"` // 所有描述对话的消息列表
}

func LoadingConfig(path, apiKey string) (Config, error) {
	if apiKey == "" {
		f, err := os.ReadFile("config/api_keys.json")
		if err != nil {
			return Config{}, err
		}
		var keys []struct{
			KeyNo string `json:"key_no"`
			Status int `json:"status"`
		}
		log.Printf("keys ： %+v", keys)
		_ = json.Unmarshal(f, &keys)
		for _, key := range keys {
			if key.Status == 0 {
				continue
			}
			apiKey = key.KeyNo
			break
		}
	}
	return Config{HOST + path,apiKey}, nil
}

func (c Config) Send(params []byte) (*http.Response, error) {
	log.Print(string(params))
	reader := bytes.NewReader(params)
	r, err := http.NewRequest(http.MethodPost, c.Uri, reader)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiKey))
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//log.Println("响应头", resp.Header)
	return resp, nil
}