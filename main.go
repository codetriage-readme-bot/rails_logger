package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"regexp"
	"strings"
)

type DataStore struct {
	Logs []*LogItem
}

type LogItem struct {
	rawData string
}

type ParsedLog struct {
	url                      string
	method                   string
	timestamp                string
	responseCode             string
	responseTime             string
	genesisTotalResponseTime string
}

func (l ParsedLog) headline() string {
	var headlineItems = make([]string, 3)
	headlineItems = append(headlineItems, fmt.Sprintf("%s\t%s\t%s", l.responseCode, l.method, l.url))
	if l.responseTime != "" {
		headlineItems = append(headlineItems, fmt.Sprintf("\nin %sms", l.responseTime))
	}
	if l.genesisTotalResponseTime != "" {
		headlineItems = append(headlineItems, fmt.Sprintf("api:%sms", l.genesisTotalResponseTime))
	}
	return strings.Join(headlineItems, " ")
}

func (l LogItem) parsedLog() ParsedLog {
	log := new(ParsedLog)
	var started_regex = regexp.MustCompile(`Started (\w+) "(.*?)" for ([a-zA-Z:0-9\.]*) at (.*)`)
	if started_regex.MatchString(l.rawData) {
		var matches = started_regex.FindStringSubmatch(l.rawData)
		log.url = matches[2]
		log.method = matches[1]
		log.timestamp = matches[4]
	}
	var completed_regex = regexp.MustCompile(`Completed\s*(?P<response_code>\w+)\s*(?P<response_description>[\w\s]+) in (?P<response_time>\d+)ms(?:\s\(Genesis: (?P<genesis_total_response_time>\d+\.\d*)ms\))?`)
	if completed_regex.MatchString(l.rawData) {
		var matches = completed_regex.FindStringSubmatch(l.rawData)
		log.responseCode = matches[1]
		log.responseTime = matches[3]
		log.genesisTotalResponseTime = matches[4]
	}
	return *log
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
		fmt.Println(log.parsedLog().headline())
	}
}
