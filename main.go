package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	qli "github.com/thbkrkr/qli/client"
)

const (
	name   = "catapi@bot"
	suffix = "-bot2"
)

var (
	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	qlient *qli.Qlient
)

func main() {
	now := time.Now()
	var err error

	qlient, err = qli.NewClientFromEnv(name)
	fatal(err, "fail to create kafka client")
	go qlient.CloseOnSig()

	topic, err := qlient.Sub()
	fatal(err, "fail to consume topic")

	log.WithFields(log.Fields{
		"name": name, "duration": time.Since(now),
	}).Info("started")

	pub(newEv("hey!"))
	pubCat()

	for msg := range topic {
		handle(msg)
	}
}

var (
	whitelist = []string{"miaou", "cat", "chat"}
	catAPI    = "http://thecatapi.com/api/images/get?format=src&type=gif"
)

type Event struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

func newEv(msg string) Event {
	return Event{
		User:    name + suffix,
		Message: msg,
	}
}

func handle(msg []byte) {
	if _, ok := filter(msg); !ok {
		return
	}
	pubCat()
}

func filter(msg []byte) (Event, bool) {
	var ev Event
	err := json.Unmarshal(msg, &ev)
	if trace(err, "fail to unmarshal event") {
		return ev, false
	}

	lowerMsg := strings.ToLower(ev.Message)
	for _, word := range whitelist {
		if strings.Contains(lowerMsg, word) {
			return ev, true
		}
	}
	return ev, false
}

func pubCat() {
	resp, err := client.Get(catAPI)
	if trace(err, "fail to request the cat API") {
		return
	}
	url := resp.Header.Get("Location")

	ok := pub(newEv(`<img src="` + url + `">`))
	if ok {
		log.WithField("url", url).Info("cat sent")
	}
}

func pub(ev Event) bool {
	msg, err := json.Marshal(ev)
	if trace(err, "fail to marshal event") {
		return false
	}

	_, _, err = qlient.Send(msg)
	if trace(err, "fail to send kafka message") {
		return false
	}
	return true
}

func contains(message string, whitelist []string) bool {
	for _, word := range whitelist {
		if strings.Contains(message, word) {
			return true
		}
	}
	return false
}

func fatal(err error, msg string) {
	if err != nil {
		log.WithError(err).Fatal(msg)
	}
}

func trace(err error, msg string) bool {
	if err != nil {
		log.WithError(err).Error(msg)
		return true
	}
	return false
}
