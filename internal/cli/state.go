package cli

import (
	"github.com/eleinah/gator/internal/config"
	"github.com/eleinah/gator/internal/database"
)

type State struct {
	Db  *database.Queries
	Cfg *config.Config
}
