package ding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Ding struct {
	webhook      string
	msgTemplate  string
	jsonTemplate map[string]interface{}
}

func NewDing(webhook string, msgTemplate string, jsonTemplate map[string]interface{}) Ding {
	return Ding{webhook, msgTemplate, jsonTemplate}
}

func (this *Ding) Send(a ...interface{}) (string, error) {
	msg := fmt.Sprintf(this.msgTemplate, a)

	j := this.jsonTemplate
	if t, ok := j["msgtype"].(string); ok {
		j[t] = msg
	} else {
		return "", fmt.Errorf("jsonTemplate::msgtype should be a string.")
	}

	b, err := json.Marshal(j)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", this.webhook, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), err
}
