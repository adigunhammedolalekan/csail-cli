package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/saas/hostgo/pkg/types"
	"io/ioutil"
	"net/http"
	"time"
)

// const serverUrl = "http://localhost:4005/api"
// const serverUrl = "https://api.hostgoapp.com/api"
const serverUrl = "https://api.csail.app/api"

type Client interface {
	Do(endpoint, method string, payload, response interface{}) error
	DoRaw(endpoint, method string, payload interface{}) ([]byte, error)
}

type defaultClient struct {
	httpClient *http.Client
	account    *types.Account
}

func NewHttpClient(account *types.Account) Client {
	return &defaultClient{account: account, httpClient: &http.Client{Timeout: 60 * time.Second}}
}

func (d *defaultClient) Do(endpoint, method string, payload, response interface{}) error {
	targetUrl := fmt.Sprintf("%s%s", serverUrl, endpoint)
	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, targetUrl, bytes.NewBuffer(p))
	if err != nil {
		return err
	}
	if d.account != nil {
		req.Header.Set("X-Auth-Token", d.account.Token)
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
		}
	}()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type serverResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	srvResponse := &serverResponse{}
	if err := json.Unmarshal(data, srvResponse); err != nil {
		return err
	}
	if code := resp.StatusCode; code != http.StatusOK || srvResponse.Error {
		return errors.New(srvResponse.Message)
	}
	if err := json.Unmarshal(data, response); err != nil {
		return err
	}
	return nil
}

func (d *defaultClient) DoRaw(endpoint, method string, payload interface{}) ([]byte, error) {
	targetUrl := fmt.Sprintf("%s%s", serverUrl, endpoint)
	req, err := http.NewRequest(method, targetUrl, nil)
	if err != nil {
		return nil, err
	}
	if d.account != nil {
		req.Header.Set("X-Auth-Token", d.account.Token)
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error. returned http code %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
