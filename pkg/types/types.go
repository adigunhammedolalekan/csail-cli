package types

type Account struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Token       string `json:"account_token"`
	CompanyName string `json:"company_name"`
}

type App struct {
	Id        uint   `json:"id"`
	AppName   string `json:"app_name"`
	AccessUrl string `json:"access_url"`
}

type Config struct {
	AppName string `yaml:"app_name"`
}

type DeploymentResult struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		AccessUrl string `json:"access_url"`
		Version string `json:"version"`
	}
}

type Env struct {
	Key string `json:"key"`
	Value string `json:"value"`
}
