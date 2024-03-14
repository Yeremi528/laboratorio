package user

import (
	"fmt"

	"github.com/Yeremi528/laboratorio/foundation/logger"
	"github.com/jmoiron/sqlx"
)

type Core struct {
	logger *logger.Logger
	db     *sqlx.DB
}

func NewCore(logger *logger.Logger, db *sqlx.DB) *Core {
	return &Core{
		logger: logger,
		db:     db,
	}
}

func (c *Core) CreateUser() {
	fmt.Printf("works")
}
