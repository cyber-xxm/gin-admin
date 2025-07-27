package jwtx

import (
	jsoniter "github.com/json-iterator/go"
)

type TokenInfo interface {
	GetAccessToken() string
	GetTokenType() string
	GetRefreshToken() string
	GetExpiresAt() int64
	EncodeToJSON() ([]byte, error)
}

type tokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int64  `json:"expires_at"`
}

func (t *tokenInfo) GetAccessToken() string {
	return t.AccessToken
}

func (t *tokenInfo) GetRefreshToken() string { return t.RefreshToken }

func (t *tokenInfo) GetTokenType() string {
	return t.TokenType
}

func (t *tokenInfo) GetExpiresAt() int64 {
	return t.ExpiresAt
}

func (t *tokenInfo) EncodeToJSON() ([]byte, error) {
	return jsoniter.Marshal(t)
}
