package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var dler Downloader

func init() {
	dler.notifier = make(chan bool)
}

type Downloader struct {
	queue []*Task
	m sync.Mutex
	notifier chan bool
}

func (dl *Downloader) Push(task *Task) {
	defer dl.m.Unlock()

	dl.m.Lock()
	dl.queue = append(dl.queue, task)

	if dl.Size() == 1 {
		log.Println("开始通知")
		dl.notifier <- true
	}
}

func (dl *Downloader) Pop() *Task {
	defer dl.m.Unlock()

	dl.m.Lock()
	if len(dl.queue) == 0 {
		return nil
	}

	task := dl.queue[0]
	dl.queue = dl.queue[1:]

	return task
}

func (dl *Downloader) Size() int {
	return len(dl.queue)
}

func errorTask(task *Task) {
	task.Status = StatusFailed
	task.FinishAt = time.Now()
	task.Save()
	return
}

func successTask(task *Task) {
	task.Status = StatusFinished
	task.FinishAt = time.Now()
	task.Save()
	return
}

func (dl *Downloader) download(task *Task) {
	resp, err := http.Get(task.URL)

	if err != nil {
		log.Println("download task error:", err)
		errorTask(task)
		return
	}

	url := strings.Split(task.URL, "?")[0]
	tmps := strings.Split(url, "/")
	filename := tmps[len(tmps)-1]

	file, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		errorTask(task)
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println(err)
		errorTask(task)
		return
	}

	cmd := exec.Command(ffsend, "upload", "-d", strconv.Itoa(task.Limit), filename)

	sout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		errorTask(task)
		return
	}

	go func() {
		buf := bufio.NewReader(sout)

		for {
			lineBytes, _, err := buf.ReadLine()
			line := string(lineBytes)
			if line == "" {
				continue
			}
			if err != nil {
				log.Println("read error:", err)
				return
			}
			task.Result = line
			log.Println("上传成功:", line)

			os.Remove(filename)
			successTask(task)
			return
		}
	}()

	err = cmd.Run()

	if err != nil {
		log.Println(err)
		errorTask(task)
		return
	}
}

func (dl *Downloader) Start() {
	go func() {
		for {
			log.Println("start notifier")
			<- dl.notifier
			log.Println("receive notifier")
			for {
				task := dl.Pop()
				if task == nil {
					break
				}

				dl.download(task)
			}
		}
	}()
}