/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

import (
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/hedzr/voxr-common/xs/mjwt"
	"github.com/labstack/echo"
	"strings"
)

func (s *echoServerImpl) initJWT(e *echo.Echo) {
	if vxconf.GetBoolR("server.jwt.enabled", true) {
		// tokenLookup := strings.Replace(vxconf.GetStringR("server.jwt.tokenLookup", ""), "?", echo.HeaderAuthorization, -1)
		fSigningKey := vxconf.GetStringR("server.jwt.signingKey", "")
		fTokenLookup := vxconf.GetStringR("server.jwt.tokenLookup", "header:?")
		fAuthScheme := vxconf.GetStringR("server.jwt.authScheme", "Bearer")
		fContextKey := vxconf.GetStringR("server.jwt.contextKey", "user")
		fTokenLookup = strings.Replace(fTokenLookup, "?", echo.HeaderAuthorization, -1)

		cfg := mjwt.JWTConfig{
			Skipper: s.cool.JWTSkipper,
			BeforeFunc: func(context echo.Context) {
				// nothing to do now
			},
			SuccessHandler: func(context echo.Context) {
				// nothing to do now
			},
			ErrorHandler: func(e error) error {
				return e
			},
			SigningKey:    fSigningKey,
			SigningMethod: vxconf.GetStringR("server.jwt.signingMethod", "HS256"),
			ContextKey:    fContextKey,
			// Claims:        &webui.JwtCustomClaims{}, //Claims jwt.Claims,
			Claims:      &mjwt.ImClaims{},
			TokenLookup: fTokenLookup, // "query:token", // tokenLookup,
			AuthScheme:  fAuthScheme,
		}

		cfg = s.cool.OnInitJWT(e, cfg)

		e.Use(mjwt.JWTWithConfig(cfg))
	}
}
