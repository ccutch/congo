package main

import (
	"io/fs"
	"log"
)

// gallery returns 4 lists of images from the gallery directory
func gallery() [4][]string {
	gallery, err := fs.Sub(public, "public/gallery")
	if err != nil {
		log.Println("Gallery not found")
		return [4][]string{}
	}

	var files []string
	fs.WalkDir(gallery, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, d.Name())
		return nil
	})

	length := len(files)
	if length == 0 {
		return [4][]string{}
	}

	result := [4][]string{}
	for i, file := range files {
		result[i%4] = append(result[i%4], file)
	}

	return result
}
