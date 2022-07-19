package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
}

func NewHTTPClient() *Client {
	return &Client{}
}

func (c *Client) Request(method string, url string, reqData interface{}, respData interface{}, headers http.Header) error {
	client := &http.Client{
		Timeout: time.Second * 100000,
	}

	var body io.Reader
	if reqData != nil {
		jd, err := json.Marshal(reqData)
		if err != nil {
			return errors.Wrap(err, "marshal request as JSON")
		}
		body = bytes.NewBuffer(jd)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	if headers != nil {
		req.Header = headers
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "request error")
	}
	defer resp.Body.Close()

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		if b, err := ioutil.ReadAll(resp.Body); err == nil && len(b) > 0 {
			fmt.Println("Response from not accepted request contain information: ", b)
		}
		return errors.Errorf("%v %v Status %v (%v)", method, url, resp.StatusCode, resp.Status)
	}

	switch respData := respData.(type) {
	case nil:
		return nil
	case *[]byte:
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
		if err := json.Unmarshal(b, respData); err != nil {
			return errors.Wrap(err, "decode response as JSON")
		}
	}

	return nil
}

func (c *Client) RequestNoLog(method string, url string, reqData interface{}, respData interface{}, headers http.Header) error {
	client := &http.Client{
		Timeout: time.Second * 10000,
	}

	var body io.Reader
	if reqData != nil {
		jd, err := json.Marshal(reqData)
		if err != nil {
			return errors.Wrap(err, "marshal request as JSON")
		}
		body = bytes.NewBuffer(jd)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	if headers != nil {
		req.Header = headers
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "request error")
	}
	defer resp.Body.Close()

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		if b, err := ioutil.ReadAll(resp.Body); err == nil && len(b) > 0 {
			fmt.Println("Response from not accepted request contain information: ", b)
		}
		return errors.Errorf("%v %v Status %v (%v)", method, url, resp.StatusCode, resp.Status)
	}

	switch respData := respData.(type) {
	case nil:
		return nil
	case *[]byte:
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
		if err := json.Unmarshal(b, respData); err != nil {
			return errors.Wrap(err, "decode response as JSON")
		}
	}

	return nil
}
