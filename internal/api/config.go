package api

import (
	"sync/atomic"

	"github.com/ihyaulhaq/go-server/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	SecretKey      string
	PolkaKey       string
}
