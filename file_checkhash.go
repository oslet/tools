package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var pathname string

func init() {
	flag.StringVar(&pathname, "path", "", "pathname need to calculate md5sum")

	flag.Usage = func() {
		fmt.Printf("Usage: %s pathname\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func printFile(ignoreDirs []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			//log.Print(err)
			return nil
		}
		if info.Name() == "test*" {
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
	flag.Parse()
	if len(pathname) == 0 {
		flag.Usage()
	}
	log.SetFlags(log.Lshortfile)
	ignoreDirs := []string{"test", "log", "Log"}
	err := filepath.Walk(pathname, printFile(ignoreDirs))
	if err != nil {
		log.Fatal(err)
	}
}
