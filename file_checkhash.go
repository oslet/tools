package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func printFile(ignoreDirs []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}
		if info.IsDir() {
			dir := filepath.Base(path)
			for _, d := range ignoreDirs {
				if d == dir {
					return filepath.SkipDir
				}
			}
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		h := md5.New()
		io.Copy(h, file)
		file.Close()
		fmt.Printf("%x %s\n", h.Sum(nil), path)

		return nil
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	dir := os.Args[1]
	ignoreDirs := []string{"test", ".hg", ".git"}
	err := filepath.Walk(dir, printFile(ignoreDirs))
	if err != nil {
		log.Fatal(err)
	}
}
