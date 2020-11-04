package main

import (
	"bufio"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func BuildOverlayDir(sourceDir string, destDir string, replace map[string]string) error {
	err := os.Chdir(sourceDir)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			err := os.MkdirAll(destDir+"/"+path, info.Mode())
			if err != nil {
				fmt.Println(err)
			}
		} else {
			if filepath.Ext(path) == ".in" {
				destFile := strings.TrimSuffix(path, ".in")
				skip := false

				sourceFD, err := os.Open(sourceDir + "/" + path)
				if err != nil {
					return err
				}

				destFD, err := os.OpenFile(destDir + "/" + destFile, os.O_RDWR|os.O_CREATE, info.Mode())
				if err != nil {
					return err
				}

				scanner := bufio.NewScanner(sourceFD)
				w := bufio.NewWriter(destFD)

				for scanner.Scan() {
					newLine := scanner.Text()

					for k, v := range replace {
						replaceString := fmt.Sprintf("@%s@", strings.ToUpper(k))
						newLine = strings.ReplaceAll(newLine, replaceString, v)
					}

					if strings.HasPrefix(newLine, "#WWEND") {
						skip = false
					} else if skip == true {

					} else if strings.HasPrefix(newLine, "#WWIFDEF ") {
						line := strings.Split(newLine, " ")
						if len(line) > 0 && line[1] != "false" {
							skip = true
						}
					} else if strings.HasPrefix(newLine, "#WWIFNDEF ") {
						line := strings.Split(newLine, " ")
						if len(line) > 0 && line[1] == "false" {
							skip = true
						}
					} else if strings.HasPrefix(newLine, "#WWIF ") {
						line := strings.Split(newLine, " ")
						if len(line) == 2 && line[1] != "false" {
							skip = true
						} else if len(line) > 3 {

							if line[2] == "==" {
								if line[1] != line[3] {
									skip = true
								}
							}
						}
					} else if strings.HasPrefix(newLine, "#WWELSE") {
						if skip == true {
							skip = false
						} else {
							skip = true
						}
					} else if strings.HasPrefix(newLine, "#WWINCLUDE ") {
						line := strings.Split(newLine, " ")
						//fmt.Printf("Including file (%s): %s\n", destDir + "/" + destFile, line[1])
						includeFD, err := os.Open(line[1])
						if err != nil {
							fmt.Printf("ERROR(os.Open): %s\n", err)
							return err
						}
						_, err = io.Copy(w, includeFD)
						if err != nil {
							fmt.Printf("ERROR(io.Copy): %s\n", err)
							return err
						}
						includeFD.Close()
					} else {
						_, err := w.WriteString(newLine + "\n")
						if err != nil {
							return err
						}
					}
				}
				w.Flush()
				sourceFD.Close()
				destFD.Close()
			} else {
				err := util.CopyFile(sourceDir+"/"+path, destDir+"/"+path)

				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
