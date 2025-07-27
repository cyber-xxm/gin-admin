package jwtx

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type Auther interface {
	// Generate a JWT (JSON Web Token) with the provided subject.
	GenerateToken(ctx context.Context, subject string) (TokenInfo, error)
	// Invalidate a token by removing it from the token store.
	DestroyToken(ctx context.Context, accessToken string) error
	// Parse the subject (or user identifier) from a given access token.
	ParseSubject(ctx context.Context, accessToken string) (string, error)
	// Release any resources held by the JWTAuth instance.
	Release(ctx context.Context) error
}

const defaultKey = "CG24SDVP8OHPK395GB5G"

var ErrInvalidToken = errors.New("Invalid token")

type options struct {
	signingMethod jwt.SigningMethod
	signingKey    []byte
	signingKey2   []byte
	keyFuncs      []func(*jwt.Token) (interface{}, error)
	expired       int
	tokenType     string
}

type Option func(*options)

func SetSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

func SetSigningKey(key, oldKey string) Option {
	return func(o *options) {
		o.signingKey = []byte(key)
		if oldKey != "" && key != oldKey {
			o.signingKey2 = []byte(oldKey)
		}
	}
}

func SetExpired(expired int) Option {
	return func(o *options) {
		o.expired = expired
	}
}

func New(store Storer, opts ...Option) Auther {
	o := options{
		tokenType:     "Bearer",
		expired:       7200,
		signingMethod: jwt.SigningMethodHS512,
		signingKey:    []byte(defaultKey),
	}

	for _, opt := range opts {
		opt(&o)
	}

	o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return o.signingKey, nil
	})

	if o.signingKey2 != nil {
		o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return o.signingKey2, nil
		})
	}

	return &JWTAuth{
		opts:  &o,
		store: store,
	}
}

type JWTAuth struct {
	opts  *options
	store Storer
}

func (a *JWTAuth) GenerateToken(ctx context.Context, subject string) (TokenInfo, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.opts.expired) * time.Second).Unix()

	accessToken := jwt.NewWithClaims(a.opts.signingMethod, &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt,
		NotBefore: now.Unix(),
		Subject:   subject,
	})

	accessTokenStr, err := accessToken.SignedString(a.opts.signingKey)
	if err != nil {
		return nil, err
	}

	// RefreshToken 设置 30 天过期
	refreshExpiredAt := now.Add(30 * 24 * time.Hour).Unix()
	refreshToken := jwt.NewWithClaims(a.opts.signingMethod, &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: refreshExpiredAt,
		NotBefore: now.Unix(),
		Subject:   subject,
	})

	refreshTokenStr, err := refreshToken.SignedString(a.opts.signingKey)
	if err != nil {
		return nil, err
	}

	tokenInfo := &tokenInfo{
		ExpiresAt:    expiresAt,
		TokenType:    a.opts.tokenType,
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
	}

	// 可选：把 refresh token 存起来用于后续校验和注销
	_ = a.callStore(func(store Storer) error {
		return store.Set(ctx, refreshTokenStr, time.Until(time.Unix(refreshExpiredAt, 0)))
	})

	return tokenInfo, nil
}

func (a *JWTAuth) RefreshToken(ctx context.Context, refreshToken string) (TokenInfo, error) {
	claims, err := a.parseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 检查是否被吊销
	err = a.callStore(func(store Storer) error {
		if exists, err := store.Check(ctx, refreshToken); err != nil {
			return err
		} else if exists {
			return ErrInvalidToken
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 再生成新的 access 和 refresh token
	return a.GenerateToken(ctx, claims.Subject)
}

func (a *JWTAuth) parseToken(tokenStr string) (*jwt.StandardClaims, error) {
	var (
		token *jwt.Token
		err   error
	)

	for _, keyFunc := range a.opts.keyFuncs {
		token, err = jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, keyFunc)
		if err != nil || token == nil || !token.Valid {
			continue
		}
		break
	}

	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return token.Claims.(*jwt.StandardClaims), nil
}

func (a *JWTAuth) callStore(fn func(Storer) error) error {
	if store := a.store; store != nil {
		return fn(store)
	}
	return nil
}

func (a *JWTAuth) DestroyToken(ctx context.Context, tokenStr string) error {
	claims, err := a.parseToken(tokenStr)
	if err != nil {
		return err
	}

	return a.callStore(func(store Storer) error {
		expired := time.Until(time.Unix(claims.ExpiresAt, 0))
		return store.Set(ctx, tokenStr, expired)
	})
}

func (a *JWTAuth) ParseSubject(ctx context.Context, tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", ErrInvalidToken
	}

	claims, err := a.parseToken(tokenStr)
	if err != nil {
		return "", err
	}

	err = a.callStore(func(store Storer) error {
		if exists, err := store.Check(ctx, tokenStr); err != nil {
			return err
		} else if exists {
			return ErrInvalidToken
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return claims.Subject, nil
}

func (a *JWTAuth) Release(ctx context.Context) error {
	return a.callStore(func(store Storer) error {
		return store.Close(ctx)
	})
}
