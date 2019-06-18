/*
 * Copyright © 2019 Hedzr Yeh.
 */

package mjwt

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// Algorithms
const (
	AlgorithmHS256    = "HS256"
	DefaultCookieName = "zag-token"
)

type (
	ImClaims struct {
		jwt.StandardClaims
		DeviceId string `json:"did" yaml:"device-id"`
	}
)

// Errors
var (
	ErrJWTMissing = echo.NewHTTPError(http.StatusBadRequest, "Missing or malformed jwt")
	ErrJWTInvalid = echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired jwt")
)

type (
	// JWTConfig defines the config for JWT middleware.
	JWTConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// BeforeFunc defines a function which is executed just before the middleware.
		BeforeFunc middleware.BeforeFunc

		// SuccessHandler defines a function which is executed for a valid token.
		SuccessHandler   JWTSuccessHandler
		ClaimsGotHandler JWTClaimsGotHandler

		// ErrorHandler defines a function which is executed for an invalid token.
		// It may be used to define a custom JWT error.
		ErrorHandler JWTErrorHandler

		// Signing key to validate token.
		// Required.
		SigningKey interface{}

		// Signing method, used to check token signing method.
		// Optional. Default value HS256.
		SigningMethod string

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string

		// Claims are extendable claims data defining token content.
		// Optional. Default value jwt.MapClaims
		Claims jwt.Claims

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup string

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string

		keyFunc jwt.Keyfunc
	}

	// JWTSuccessHandler defines a function which is executed for a valid token.
	JWTSuccessHandler func(echo.Context)

	JWTClaimsGotHandler func(c echo.Context, tokenString string, token *jwt.Token, imtk *ImClaims)

	// JWTErrorHandler defines a function which is executed for an invalid token.
	JWTErrorHandler func(error) error

	jwtExtractor func(echo.Context) (string, error)
)

var (
	// DefaultJWTConfig is the default JWT auth middleware config.
	DefaultJWTConfig = JWTConfig{
		Skipper:       middleware.DefaultSkipper,
		SigningMethod: AlgorithmHS256,
		ContextKey:    "user",
		TokenLookup:   "header:" + echo.HeaderAuthorization,
		AuthScheme:    "Bearer",
		Claims:        jwt.MapClaims{},
	}
)

// JWT returns a JSON Web Token (JWT) auth middleware.
//
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it returns "401 - Unauthorized" error.
// For missing token, it returns "400 - Bad Request" error.
//
// See: https://jwt.io/introduction
// See `JWTConfig.TokenLookup`
func JWT(key interface{}) echo.MiddlewareFunc {
	c := DefaultJWTConfig
	c.SigningKey = key
	return JWTWithConfig(c)
}

// var extractor2 jwtExtractor

// JWTWithConfig returns a JWT auth middleware with config.
// See: `JWT()`.
func JWTWithConfig(config JWTConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}
	if config.SigningKey == nil {
		panic("echo: jwt middleware requires signing key")
	}
	if s, ok := config.SigningKey.(string); ok {
		config.SigningKey = []byte(s)
	}
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultJWTConfig.AuthScheme
	}
	config.keyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return config.SigningKey, nil
	}

	// Initialize
	var extractors []jwtExtractor
	parts := strings.Split(config.TokenLookup, ":")
	switch parts[0] {
	case "query":
		extractors = append(extractors, jwtFromQuery(parts[1]))
		extractors = append(extractors, jwtFromCookie(DefaultCookieName))
	case "cookie":
		extractors = append(extractors, jwtFromCookie(parts[1]))
	default:
		extractors = append(extractors, jwtFromHeader(parts[1], config.AuthScheme))
		extractors = append(extractors, jwtFromCookie(DefaultCookieName))
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var (
				auth string
				err  error
			)
			if config.Skipper(c) {
				// auth, err = extractor(c)
				// if err != nil {
				return next(c)
				// }
			}

			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			if err == nil && len(auth) == 0 {
				for _, extractor := range extractors {
					auth, err = extractor(c)
					if err == nil {
						break
					}
				}
				if err != nil {
					if config.ErrorHandler != nil {
						return config.ErrorHandler(err)
					}
					return err
				}
			}
			token := new(jwt.Token)
			// Issue #647, #656
			if _, ok := config.Claims.(jwt.MapClaims); ok {
				token, err = jwt.Parse(auth, config.keyFunc)

			} else {
				t := reflect.ValueOf(config.Claims).Type().Elem()
				claims := reflect.New(t).Interface().(jwt.Claims)
				token, err = jwt.ParseWithClaims(auth, claims, config.keyFunc)

				logrus.Debugf("    - jwt . decoded: custom claims = %v, err = %v, user-agent: %v", claims, err, c.Request().UserAgent())
				if imtk, ok := claims.(*ImClaims); ok {
					// search redis cache for the token string.
					logrus.Debugf("    - jwt . imtk: uid=%v, did=%v", imtk.Id, imtk.DeviceId)
					if config.ClaimsGotHandler != nil {
						config.ClaimsGotHandler(c, auth, token, imtk)
					}
				}
			}
			if err == nil && token.Valid {
				// Store user information from token into context.
				c.Set(config.ContextKey, token)
				if config.SuccessHandler != nil {
					config.SuccessHandler(c)
				}
				return next(c)
			}
			if config.ErrorHandler != nil {
				return config.ErrorHandler(err)
			}
			return &echo.HTTPError{
				Code:     http.StatusUnauthorized,
				Message:  "invalid or expired jwt",
				Internal: err,
			}
		}
	}
}

// var extractors []jwtExtractor

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		auth := c.Request().Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", ErrJWTMissing
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the query string.
func jwtFromQuery(param string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", ErrJWTMissing
		}
		return token, nil
	}
}

// jwtFromCookie returns a `jwtExtractor` that extracts token from the named cookie.
func jwtFromCookie(name string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		cookie, err := c.Cookie(name)
		if err != nil {
			return "", ErrJWTMissing
		}
		return cookie.Value, nil
	}
}
