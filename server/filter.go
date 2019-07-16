package server

import (
	"github.com/astaxie/beego/context"
	"github.com/yushizhao/authenticator/boltwrapper"
	"github.com/yushizhao/authenticator/gawrapper"
	"github.com/yushizhao/authenticator/jwtwrapper"
	"github.com/yushizhao/hubble/config"

	"github.com/astaxie/beego/plugins/cors"
)

var FilterCrossDomain = cors.Allow(&cors.Options{
	AllowAllOrigins:  true,
	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
	ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
	AllowCredentials: true,
})

var FilterJWT = func(ctx *context.Context) {
	token := ctx.Input.Header("X-JWT")
	if token == "" {
		ctx.Abort(401, "Missing Token")
	}

	claims, err := jwtwrapper.GetMapClaims(token, config.Conf.JWTSecret)
	if err != nil {
		ctx.Abort(401, err.Error())
	}

	name, ok := claims["Name"].(string)
	if !ok {
		ctx.Abort(401, "Invalid Token Context")
	}

	userBytes := boltwrapper.UserDB.GetUser(name)
	if userBytes == nil {
		ctx.Abort(401, "UserName Not Exists")
	}
}

var FilterRootTOTP = func(ctx *context.Context) {
	yourCode := ctx.Input.Query("Root")
	if yourCode == "" {
		ctx.Abort(401, "Missing Root")
	}

	verified, err := gawrapper.VerifyTOTP(config.Conf.RootKey, yourCode)
	if err != nil {
		ctx.Abort(500, err.Error())
	}
	if !verified {
		ctx.Abort(401, "Invalid Root")
	}
}
