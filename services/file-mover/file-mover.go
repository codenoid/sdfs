package main

import (
	"fmt"
	"path/filepath"
	"os"
	"strings"
)

import (
	"../../master-server/helper"
)

func main() {

	if masterPath := os.Args[1:]; len(masterPath) == 1 {
		// implement lazy-mode check path please contribute to
		// check whenever path is a folder or a file
		// or just remove the path check, so there is will be error
		// if the given path doesn't exist
		if _, err := os.Stat(masterPath[0]); os.IsNotExist(err) {
			// path/to/whatever does not exist
			fmt.Fprintf(os.Stderr, "error: path doesn't exist")
			os.Exit(1)
		}

		brick, err := helper.AvailableBrick()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		e := filepath.Walk(masterPath[0], func(path string, info os.FileInfo, err error) error {
			if err == nil {

				if info.IsDir() == false {
					// check for file status (symlink, etc)
					// https://linux.die.net/man/2/lstat
					fi, err := os.Lstat(path)

					if err == nil {
						if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
							// symlinked file
							// :future use
						} else {

							// not safe mf
							originPath := strings.Replace(path, masterPath[0], "", 1)
							brickPath := brick + originPath

							dirOnly := strings.Split(brickPath, "/")
							dirOnly = dirOnly[:len(dirOnly)-1]

							os.MkdirAll(strings.Join(dirOnly[:], "/"), os.ModePerm)

							err = helper.MoveFile(path, brickPath)

							if err == nil {
								helper.Symlink(brickPath, path)
							}
						}
					}
				}
			}
			return nil
		})
		if e != nil {
			fmt.Println(e)
		}

	}
}
