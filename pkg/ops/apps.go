package ops

import (
	"fmt"
	"github.com/saas/hostgo/pkg/http"
	"github.com/saas/hostgo/pkg/types"
)

type AppsOp struct {
	httpClient http.Client
}

func NewAppsOp(httpClient http.Client) *AppsOp {
	return &AppsOp{
		httpClient: httpClient,
	}
}

func (op *AppsOp) CreateNewApp(name string) (*types.App, error) {
	type payload struct {
		Name string `json:"name"`
	}
	p := &payload{Name: name}
	type serverResponse struct {
		Error   bool       `json:"error"`
		Message string     `json:"message"`
		Data    *types.App `json:"data"`
	}
	s := &serverResponse{}
	err := op.httpClient.Do("/me/apps", "POST", p, s)
	if err != nil {
		return nil, err
	}
	return s.Data, nil
}

func (op *AppsOp) ReadLogs(appName string) (string, error) {
	type serverResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    struct {
			Logs string `json:"logs"`
		} `json:"data"`
	}
	s := &serverResponse{}
	err := op.httpClient.Do(fmt.Sprintf("/apps/logs/%s", appName), "GET", nil, s)
	if err != nil {
		return "", err
	}
	return s.Data.Logs, nil
}
