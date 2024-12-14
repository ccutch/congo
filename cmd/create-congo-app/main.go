package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ccutch/congo/pkg/congo_code"
)

func main() {
	var (
		dest = flag.String("dest", "", "The destination directory where files will be generated.")
		name = flag.String("name", "", "The namespace to interpolate into the templates.")
	)

	flag.Parse()

	// Validate required flags
	if *name == "" {
		fmt.Println("Usage: generator -name=<namespace> (-dest=<destination-directory>)")
		flag.PrintDefaults()
		os.Exit(1)
	} else if *dest == "" {
		*dest = "./" + *name
	}

	// Run the generator
	log.Printf("Generating files in '%s' with namespace '%s'...", *dest, *name)
	err := congo_code.GenerateExample(*dest, *name)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Install dependencies
	run("cd", *dest, "&&", "go", "mod", "tidy")
	run("cd", *dest, "&&", "go", "get", "-u", "github.com/ccutch/congo@latest")

	// Pring next steps for user
	log.Println("Files generated successfully.")
	fmt.Print("Run the following commands:\n\n")
	fmt.Println(" $ cd ", *dest)
	fmt.Print(" $ go run .\n\n")
}

func run(args ...string) (stdout bytes.Buffer, stderr bytes.Buffer) {
	if len(args) == 0 {
		log.Fatal("missing arguments")
	}

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Fatal("failed to run cmd: ", err)
	}

	return
}
