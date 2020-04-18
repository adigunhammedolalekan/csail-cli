package ops

import (
	"fmt"
	"github.com/saas/hostgo/pkg/http"
	"github.com/saas/hostgo/pkg/types"
	"io/ioutil"
	"os"
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

func (op *AppsOp) DockerDeploy(appName, dockerUrl string) (string, error) {
	type serverResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data struct{
			Address string `json:"address"`
			Version string `json:"version"`
		} `json:"data"`
	}
	type payload struct {
		AppName string `json:"app_name"`
		DockerUrl string `json:"docker_url"`
	}
	p := &payload{AppName: appName, DockerUrl: dockerUrl}
	s := &serverResponse{}
	err := op.httpClient.Do("/apps/docker/deploy", "POST", p, s)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s | %s", s.Data.Address, s.Data.Version), nil
}

func (op *AppsOp) AddDomain(appName, domain string) (string, error) {
	type serverResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	type payload struct {
		AppName string `json:"app_name"`
		Domain string `json:"domain"`
	}
	p := &payload{AppName: appName, Domain: domain}
	s := &serverResponse{}
	err := op.httpClient.Do("/apps/domain/new", "POST", p, s)
	if err != nil {
		return "", err
	}
	return "domain added successfully", nil
}

func (op *AppsOp) RemoveDomain(appName, domain string) (string, error) {
	type serverResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	type payload struct {
		AppName string `json:"app_name"`
		Domain string `json:"domain"`
	}
	p := &payload{AppName: appName, Domain: domain}
	s := &serverResponse{}
	err := op.httpClient.Do("/apps/domain/remove", "DELETE", p, s)
	if err != nil {
		return "", err
	}
	return "domain removed successfully", nil
}

func (op *AppsOp) DumpDatabase(appName, resName string) (string, error) {
	data, err := op.httpClient.DoRaw(
		fmt.Sprintf("/apps/resource/dump/%s?res=%s", appName, resName),
		"GET", nil)
	if err != nil {
		return "", err
	}
	destination := fmt.Sprintf("%s-%s.sql", appName, resName)
	err = ioutil.WriteFile(destination, data, os.ModePerm)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("database successfully dumped to %s.", destination), nil
}