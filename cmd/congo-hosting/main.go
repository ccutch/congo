package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatal("Missing command: launch, connect, restart, help")
	}
	switch cmd := flag.Arg(0); cmd {
	case "launch":
		if server, err := launch(flag.Args()...); err != nil {
			log.Fatal("Failed to launch server:", err)
		} else {
			log.Println("Successfully launched:", server.Server.Addr())
		}

	case "restart":
		if err := restart(flag.Args()...); err != nil {
			log.Fatal("Failed to restart to server:", err)
		}

	case "connect":
		if err := connect(flag.Args()...); err != nil {
			log.Fatal("Failed to connect to server:", err)
		}

	case "gen-certs":
		if err := genCerts(flag.Args()...); err != nil {
			log.Fatal("Failed to generate certs:", err)
		}

	case "destroy":
		if err := destroy(flag.Args()...); err != nil {
			log.Fatalf("Failed to destroy server: %v", err)
		}

	case "help":
		flag.Usage()

	default:
		log.Println("Unknown Command:", cmd)
		fmt.Println()
		flag.Usage()
		fmt.Println()
	}
}
