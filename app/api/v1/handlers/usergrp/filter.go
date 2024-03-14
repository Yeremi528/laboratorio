package usergrp

import (
	"github.com/Yeremi528/laboratorio/business/core/user"
	"github.com/Yeremi528/laboratorio/foundation/logger"
	"github.com/Yeremi528/laboratorio/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log *logger.Logger
	DB  *sqlx.DB
}

// Route adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"
	usrCore := user.NewCore(cfg.Log, cfg.DB)

	New(usrCore)
	//app.Handle(http.MethodGet, version, "/users", hdl.query)

}
