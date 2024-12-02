package monitoring

import (
	"embed"
	"log"
	"os"
	"time"

	"github.com/ccutch/congo/pkg/congo"
)

//go:embed all:migrations
var migrations embed.FS

func Start(server *congo.Server) {
	root := os.Getenv("DATA_PATH")
	if root == "" {
		log.Println("[MONITOR] $DATA_PATH not set. Not monitoring")
		return
	}

	db := congo.SetupDatabase(root, "_monitor_.sqlite", migrations)
	defer db.Close()

	if err := db.MigrateUp(); err != nil {
		log.Println("[MONITOR] Failed to setup database:", err)
		return
	}

	server.WithEndpoint("/_monitor/", false, DisplaySystemMetrics)

	for {
		time.Sleep(5 * time.Second)
		if status, err := GetSystemStatus(db); err != nil {
			log.Println("[MONITOR] Failed to get system status:", err)
		} else if err := status.Save(); err != nil {
			log.Println("[MONITOR] Failed saving system status:", err)
		}
	}
}
