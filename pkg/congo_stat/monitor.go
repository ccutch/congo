package congo_stat

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

type Monitor struct {
	app *congo.Application
	dir *congo_auth.Directory
}

func NewMonitor(app *congo.Application, dir *congo_auth.Directory) *Monitor {
	return &Monitor{app, dir}
}

func (m *Monitor) Start() error {
	root := os.Getenv("DATA_PATH")
	if root == "" {
		log.Println("[MONITOR] $DATA_PATH not set. Not monitoring")
		return nil
	}

	db := congo.SetupDatabase(root, "monitor.db", migrations)
	defer db.Close()

	if err := db.MigrateUp(); err != nil {
		log.Println("[MONITOR] Failed to setup database:", err)
		return err
	}

	if m.dir != nil {
		m.app.HandleFunc("/_stat/", m.dir.SecureFunc(m.viewStatusHistory))
	} else {
		log.Println("Provide an auth directory to secure")
	}

	for {
		time.Sleep(5 * time.Second)
		if status, err := GetSystemStatus(db); err != nil {
			log.Println("[MONITOR] Failed to get system status:", err)
		} else if err := status.Save(); err != nil {
			log.Println("[MONITOR] Failed saving system status:", err)
		}
	}
}
