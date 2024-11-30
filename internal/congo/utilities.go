package congo

import "log"

func Must(err error) {
	if err != nil {
		log.Fatal("Failed to setup Congo server:", err)
	}
}
