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

func (op *AppsOp) RollbackDeployment(appName, version string) (string, error) {
	type serverResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    struct {
			Address string `json:"address"`
			Version string `json:"version"`
		} `json:"data"`
	}
	s := &serverResponse{}
	err := op.httpClient.Do(fmt.Sprintf("/apps/rollback/%s?version=%s", appName, version), "PUT", nil, s)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s | %s", s.Message, s.Data.Version), nil
}

func (op *AppsOp) ProvisionResource(appName, resName string) (string, error) {
	type serverResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    struct {
			Id string `json:"id"`
		} `json:"data"`
	}
	s := &serverResponse{}
	err := op.httpClient.Do(fmt.Sprintf("/apps/resource/new/%s?name=%s", appName, resName), "POST", nil, s)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s | %s", s.Message, s.Data.Id), nil
}

func (op *AppsOp) DeleteResource(appName, resName string) (string, error) {
	type serverResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	s := &serverResponse{}
	err := op.httpClient.Do(fmt.Sprintf("/apps/resource/remove/%s?name=%s", appName, resName), "DELETE", nil, s)
	if err != nil {
		return "", err
	}
	return s.Message, nil
}
