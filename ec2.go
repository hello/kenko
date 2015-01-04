package main

import (
	"io/ioutil"
	"net/http"
	"time"
)

type httpResponse struct {
	url      string
	response *http.Response
	err      error
}

func GetTags() []string {
	tags := make([]string, 0)
	urls := []string{"placement/availability-zone", "instance-id"}
	resps := asyncHttpGets(urls, time.Duration(1000))
	for _, resp := range resps {
		buff, err := ioutil.ReadAll(resp.response.Body)
		if err == nil {
			tags = append(tags, string(buff))
		}
	}
	return tags
}

func asyncHttpGets(urls []string, timeoutMillis time.Duration) []*httpResponse {
	ch := make(chan *httpResponse)
	responses := []*httpResponse{}
	for _, url := range urls {
		go func(url string) {
			resp, err := http.Get("http://169.254.169.254/latest/meta-data/" + url)
			ch <- &httpResponse{url, resp, err}
		}(url)
	}

	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == len(urls) {
				return responses
			}
		case <-time.After(timeoutMillis * time.Millisecond):
			return responses
		}
	}
	return responses
}
