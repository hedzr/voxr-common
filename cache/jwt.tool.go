/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cache

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/hedzr/voxr-common/xs/mjwt"
	"github.com/labstack/echo"
	"time"
)

const (
	AlgorithmHS256    = "HS256"
	DefaultSigningKey = "fxxking-secrets-here"
)

var (
	signingMethod string
	signingKey    string
	keyFunc       jwt.Keyfunc
)

func JwtInit() {
	signingMethod = vxconf.GetStringR("server.jwt.signingMethod", AlgorithmHS256)
	signingKey = vxconf.GetStringR("server.jwt.signingKey", DefaultSigningKey)
	keyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != signingMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return []byte(signingKey), nil
	}
}

func JwtSign(userId, deviceId string) (string, error) {
	now := time.Now().UnixNano()

	// Create the Claims
	claims := &mjwt.ImClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Unix(0, now).Add(time.Hour * 24 * 7).Unix(),
			Issuer:    "im-auth",
			Id:        userId, // user id here
			Subject:   ">",
		},
		DeviceId: deviceId,
	}

	// JwtInit()

	token := jwt.NewWithClaims(jwt.GetSigningMethod(signingMethod), claims)
	ss, err := token.SignedString([]byte(signingKey))
	fmt.Printf("%v %v", ss, err)
	if err != nil {
		return "", err
	}
	return ss, nil
}

func JwtVerifyToken(tokenString string) (valid bool, err error) {
	var token *jwt.Token
	token, err = jwt.Parse(tokenString, keyFunc)

	if token.Valid {
		valid = true
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			fmt.Println("That's not even a token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			fmt.Println("Timing is everything")
		} else {
			fmt.Println("Couldn't handle this token:", err)
		}
	} else {
		fmt.Println("Couldn't handle this token:", err)
	}
	return
}

func JwtExtractToken(tokenString string) (token *jwt.Token, err error) {
	token, err = jwt.ParseWithClaims(tokenString, &mjwt.ImClaims{}, keyFunc)
	return
}

func JwtExtract(c echo.Context) (token *jwt.Token, valid bool, err error) {
	if tk, ok := c.Get(vxconf.GetStringR("server.jwt.contextKey", "user")).(*jwt.Token); ok {
		token = tk
		valid = token.Valid
	} else {
		err = errors.New("invalid jwt token in context")
	}
	return
}

func JwtDecodeToken(tokenString string) (token *jwt.Token, valid bool, err error) {
	token, err = jwt.Parse(tokenString, keyFunc)

	if token.Valid {
		valid = true
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			fmt.Println("That's not even a token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			fmt.Println("Timing is everything")
		} else {
			fmt.Println("Couldn't handle this token:", err)
		}
	} else {
		fmt.Println("Couldn't handle this token:", err)
	}
	return
}

func jwtDecode(token string) (tk *jwt.Token) {
	tk, _, _ = JwtDecodeToken(token)
	return
}
