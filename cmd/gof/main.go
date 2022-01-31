package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cookie04DE/gof"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Please specify at least one file or directory with .gof files to convert\n")
	}
	for _, arg := range os.Args[1:] {
		processFile(arg, true)
	}
}

const generationWarning = "//Auto generated by gof; DO NOT EDIT\n"

func processFile(path string, toplevel bool) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error opening file %s: %s", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		log.Fatalf("Error getting file info from file %s: %s", path, err)
	}
	if !info.IsDir() {
		ext := filepath.Ext(info.Name())
		if ext != ".gof" {
			if toplevel {
				log.Fatalf("Wrong extension %q for file %s; need .gof}", ext, path)
			}
			return
		}
		targetPath := strings.TrimSuffix(path, ext) + ".go"
		targetF, err := os.Create(targetPath)
		if err != nil {
			log.Fatalf("Error creating file %s: %s", targetPath, err)
		}
		defer targetF.Close()
		_, err = targetF.WriteString(generationWarning)
		if err != nil {
			log.Fatalf("Error writing to file %s: %s", targetPath, err)
		}
		gof.Convert(f, targetF)
		return
	}
	files, err := f.Readdirnames(-1)
	if err != nil {
		log.Fatalf("Error getting filenames from directory %s: %s", path, err)
	}
	for _, file := range files {
		processFile(file, false)
	}
}