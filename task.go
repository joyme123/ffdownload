package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

var TaskSavePath = "./data"

const (
	StatusQueued = "queued"
	StatusStarted = "started"
	StatusFinished = "finished"
	StatusFailed = "failed"
)


type Task struct {
	ID string `json:"id"`
	URL string `json:"url"`
	Limit int `json:"limit"`
	Result string `json:"result"`
	Status string 	`json:"status"`
	Progress int	`json:"progress"`
	CreateAt time.Time	`json:"createAt"`
	UpdateAt time.Time	`json:"updateAt"`
	FinishAt time.Time	`json:"finishAt"`
}

func (task *Task) Save() error {

	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(TaskSavePath+"/"+task.ID, data, os.ModePerm)
}

func (task *Task) Load() error {
	data, err := ioutil.ReadFile(TaskSavePath+"/"+task.ID)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, task)
	if err != nil {
		return err
	}
	return nil
}