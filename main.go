package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"openai-go/chatGPT"
)

func main() {
	http.HandleFunc("/", sendHandle)
	if err := http.ListenAndServe(":8686", nil); err != nil {
		log.Fatal(err)
	}
}

func sendHandle(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("Api-Key")
	if apiKey == "" {
		response(w, http.StatusBadRequest, -1, "参数错误")
		return
	}
	c, err := chatGPT.LoadingConfig(r.RequestURI, apiKey)
	if err != nil {
		response(w, http.StatusBadRequest, -1, err.Error())
		return
	}
	log.Printf("gpt配置:%+v", c)

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	resp, err := c.Send(bodyBytes)
	if err != nil {
		response(w, http.StatusBadRequest, -1, "请求失败")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		aiBody := make([]byte, resp.ContentLength)
		resp.Body.Read(aiBody)
		log.Printf("失败返回参数 %s", string(aiBody))
		response(w, http.StatusBadRequest, -1, "请求失败")
		return
	}
	log.Printf("响应头 %+v", resp.Header)
	respContentType := resp.Header.Get("Content-Type")
	//w.Header().Set("Content-Type", "text/event-stream")
	// 流的形式
	if respContentType == "text/event-stream" {
		w.Header().Set("Content-Type", respContentType)
		//log.Println("响应头", resp.Header)
		//scanner := bufio.NewScanner(resp.Body)
		reader := bufio.NewReaderSize(resp.Body, 10000)
		//for scanner.Scan() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println("请求不完整")
				break
			}
			//log.Println("接收参数 ：" + line)
			//if ! scanner.Scan() {
			//	continue
			//}
			//go log.Println("接收参数 ：" + string(scanner.Bytes()))

			//if parseEventStreamFields(scanner.Text()) == "[DONE]" {
			//	break
			//}
			if line == "" {
				continue
			}
			fmt.Fprint(w, parseEventStreamFields(line))
			w.(http.Flusher).Flush()
		}
		return
	} else {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(bodyBytes))
		return
	}
}

func parseEventStreamFields(line string) string {
	field := strings.Split(line, ": ")
	if len(field) > 1 {
		return field[1] + "\n"
	}
	return "\n"
}

func response(w http.ResponseWriter, requestStatus, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	type data struct {
		Status int `json:"status"`
		Msg string 	`json:"msg"`
		Data interface{} `json:"data"`
	}
	d, _ := json.Marshal(data{status, msg, ""})
	w.WriteHeader(requestStatus)
	fmt.Fprint(w, string(d))
}