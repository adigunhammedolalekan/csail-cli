package auth

import (
	"encoding/json"
	"errors"
	"github.com/mitchellh/go-homedir"
	"github.com/saas/hostgo/pkg/types"
	"io/ioutil"
	"os"
	"path/filepath"
)

var ErrNoHomeDir = errors.New("failed to determine home directory. Please set HOME_DIR env variable to your home directory")
var ErrNoAuth = errors.New("hostGo authentication required. Please run `hostgo login` to authenticate your account")
var ErrInvalidAuthData = errors.New("failed to create auth: invalid authentication data")

const authFileName = ".auth.json"
const authDirName = ".hostgo"

type AuthProvider interface {
	CreateAuth(account *types.Account) error
	CurrentAuth() (*types.Account, error)
}

type defaultAuthProvider struct {
}

func NewAuthProvider() AuthProvider {
	return &defaultAuthProvider{}
}

func (d *defaultAuthProvider) CurrentAuth() (*types.Account, error) {
	h, err := homedir.Dir()
	if err != nil {
		h = os.Getenv("HOME_DIR")
		if h == "" {
			return nil, ErrNoHomeDir
		}
	}
	authFile := filepath.Join(h, authDirName, authFileName)
	data, err := ioutil.ReadFile(authFile)
	if err != nil {
		return nil, ErrNoAuth
	}
	account := &types.Account{}
	if err := json.Unmarshal(data, account); err != nil {
		return nil, ErrNoAuth
	}
	return account, nil
}

func (d *defaultAuthProvider) CreateAuth(account *types.Account) error {
	h, err := homedir.Dir()
	if err != nil {
		h = os.Getenv("HOME_DIR")
		if h == "" {
			return ErrNoHomeDir
		}
	}
	authFolder := filepath.Join(h, authDirName)
	if _, err := os.Stat(authFolder); err == nil {
		if err := os.RemoveAll(authFolder); err != nil {
			return errors.New("failed to authenticate account: " + err.Error())
		}
	}
	if err := os.Mkdir(authFolder, os.ModePerm); err != nil {
		return errors.New("failed to authenticate account: " + err.Error())
	}
	data, err := json.Marshal(account)
	if err != nil {
		return ErrInvalidAuthData
	}
	authFile := filepath.Join(authFolder, authFileName)
	if err := ioutil.WriteFile(authFile, data, os.ModePerm); err != nil {
		return errors.New("failed to authenticate account: " + err.Error())
	}
	return nil
}
