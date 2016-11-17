package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"regexp"
)

type DataStore struct {
	Logs []*LogItem
}

type LogItem struct {
	rawData string
}

var dataStore DataStore

var done = make(chan bool)
var logItems = make(chan LogItem)

func main() {
	go parser()
	go echoer()
	<-done
}

func parser() {
	t, err := tail.TailFile("./sample.log", tail.Config{Follow: true})
	if err != nil {
		panic(err)
	}
	logItem := new(LogItem)
	for line := range t.Lines {
		match, _ := regexp.MatchString("Started ", line.Text)
		if match {
			if logItem.rawData != "" {
				logItems <- *logItem
			}
			logItem = new(LogItem)
			logItem.rawData = line.Text
		} else {
			logItem.rawData += "\n"
			logItem.rawData += line.Text
		}
	}

	done <- true
}

func echoer() {
	for {
		log := <-logItems
		fmt.Println(log.rawData)
		fmt.Println("##########################\n\n")
	}
}
