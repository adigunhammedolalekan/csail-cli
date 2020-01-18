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

// const serverUrl = "http://localhost:4001/api"
const serverUrl = "http://167.172.159.245:4005/api"

type Client interface {
	Do(endpoint, method string, payload interface{}, response interface{}) error
}

type defaultClient struct {
	httpClient *http.Client
	account    *types.Account
}

func NewHttpClient(account *types.Account) Client {
	return &defaultClient{account: account, httpClient: &http.Client{Timeout: 60 * time.Second}}
}

func (d *defaultClient) Do(endpoint, method string, payload interface{}, response interface{}) error {
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
