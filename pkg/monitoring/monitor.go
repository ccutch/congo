package monitoring

import (
	"embed"
	"log"
	"os"
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

//go:embed all:migrations
var migrations embed.FS

func Start(app *congo.Application, auth *congo_auth.Authenticator) {
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

	if auth != nil {
		log.Println("Secure this route pls")
	}

	app.HandleFunc("/_monitor/", DisplaySystemMetrics)

	for {
		time.Sleep(5 * time.Second)
		if status, err := GetSystemStatus(db); err != nil {
			log.Println("[MONITOR] Failed to get system status:", err)
		} else if err := status.Save(); err != nil {
			log.Println("[MONITOR] Failed saving system status:", err)
		}
	}
}
