package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ccutch/congo/pkg/generator"
)

func main() {
	var (
		dest = flag.String("dest", "", "The destination directory where files will be generated.")
		name = flag.String("name", "", "The namespace to interpolate into the templates.")
	)
	flag.Parse()

	// Validate required flags
	if *name == "" {
		fmt.Println("Usage: generator -dest=<destination-directory> -name=<namespace>")
		flag.PrintDefaults()
		os.Exit(1)
	} else if *dest == "" {
		*dest = "./" + *name
	}

	// Run the generator
	log.Printf("Generating files in '%s' with namespace '%s'...", *dest, *name)
	err := generator.GenerateExample(*dest, *name)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Files generated successfully.")
}
