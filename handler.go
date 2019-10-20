package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var ffsend = "ffsend"
var path = "./assets"

type Resp struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	Data  map[string]string `json:"data"`
}

func writeErrorResponse(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	
	resp := Resp{
		Error: "1",
		Msg:   msg,
		Data:  map[string]string{},
	}

	d, err := json.Marshal(resp)
	if err != nil {
		log.Println("json marshal error", err)
		return
	}

	_, err = w.Write(d)
	if err != nil {
		log.Println("write data error:", err)
	}
}

func writeOkResponse(w http.ResponseWriter, out map[string]string) {
	resp := Resp{
		Error: "0",
		Msg:   "",
		Data:  out,
	}

	d, err := json.Marshal(resp)
	if err != nil {
		log.Println("json marshal error:", err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(d)
	if err != nil {
		log.Println("write data error:", err)
	}
}

func randString(n int) string {
	rand.Seed(time.Now().Unix())
	text := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	l := len(text)

	buf := bytes.NewBuffer([]byte{})

	for i := 0; i < n; i++ {
		p := rand.Intn(l)
		buf.WriteByte(text[p])
	}

	return buf.String()
}

func Download(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	limitStr := req.PostFormValue("limit")

	var limit int
	if limitStr == "" {
		limit = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "下载参数limit错误")
		return
	}

	if limit == 0 {
		limit = 1
	}

	taskUrl := req.PostFormValue("task")

	if taskUrl == "" {
		writeErrorResponse(w, http.StatusBadRequest, "缺少必要的下载参数")
		return
	}

	log.Println("开始任务:", taskUrl)

	taskID := randString(32)
	
	task := &Task{
		ID:       taskID,
		URL:      taskUrl,
		Status:   StatusQueued,
		Limit:    limit,
		Progress: 0,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
		FinishAt: time.Now(),
	}

	dler.Push(task)
	task.Save()

	writeOkResponse(w, map[string]string{
		"taskID": taskID,
	})

}

func Retrieve(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	taskIDs, ok := req.URL.Query()["taskID"]
	if !ok || len(taskIDs) == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "缺少taskID")
		return
	}

	taskID := taskIDs[0]
	task := &Task{
		ID:       taskID,
	}

	err := task.Load()
	if err != nil {
		log.Println("load task error:", err)
		writeErrorResponse(w, http.StatusBadRequest, "task信息获取出错")
		return
	}

	writeOkResponse(w, map[string]string{
		"status": task.Status,
		"result": task.Result,
	})
}

func main() {
	pathEnv := os.Getenv("UI_PATH")
	if pathEnv != "" {
		path = pathEnv
	}

	ffsendEnv := os.Getenv("FFSEND")
	if ffsendEnv != "" {
		ffsend = ffsendEnv
	}

	// 开启下载队列
	dler.Start()

	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)

	http.HandleFunc("/v1/api/download", Download)
	http.HandleFunc("/v1/api/retrieve", Retrieve)

	log.Println("starting on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}