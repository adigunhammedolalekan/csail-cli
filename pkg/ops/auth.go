package ops

import (
	"github.com/saas/hostgo/pkg/auth"
	"github.com/saas/hostgo/pkg/http"
	"github.com/saas/hostgo/pkg/types"
)

type AuthenticateAccountOp struct {
	client   http.Client
	provider auth.AuthProvider
}

func NewAuthenticateAccountOp(client http.Client, provider auth.AuthProvider) *AuthenticateAccountOp {
	return &AuthenticateAccountOp{client: client, provider: provider}
}

func (op *AuthenticateAccountOp) AuthenticateAccount(email, password string) (*types.Account, error) {
	type payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		Error   bool           `json:"error"`
		Message string         `json:"message"`
		Data    *types.Account `json:"data"`
	}
	p := &payload{Email: email, Password: password}
	r := &response{}
	err := op.client.Do("/account/authenticate", "POST", p, r)
	if err != nil {
		return nil, err
	}
	if err := op.provider.CreateAuth(r.Data); err != nil {
		return nil, err
	}
	return r.Data, nil
}
