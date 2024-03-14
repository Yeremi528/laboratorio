package usergrp

import (
	"net/http"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log *logger.Logger
	DB  *sqlx.Db
}

// Route adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	usrCore := user.NewCore(cfg.Log, cfg.DB)

	hdl := new(usrCore)
	app.Handle(http.MethodGet, version, "/users", hdl.query)

}
