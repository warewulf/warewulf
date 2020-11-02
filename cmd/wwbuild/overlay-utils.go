package main

import (
	"bufio"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
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
			err := os.MkdirAll(destDir + "/" + path, info.Mode())
			if err != nil {
				fmt.Println(err)
			}
		} else {
			if filepath.Ext(path) == ".in" {
				destFile := strings.TrimSuffix(path, ".in")

				sourceFD, err := os.Open(sourceDir + "/" + path)
				if err != nil {
					return err
				}

				destFD, err :=os.OpenFile(destDir + "/" + destFile, os.O_RDWR|os.O_CREATE, info.Mode())
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
					_, err := w.WriteString(newLine + "\n")
					if err != nil {
						return err
					}

				}
				w.Flush()
				sourceFD.Close()
				destFD.Close()
			} else {
				err := util.CopyFile(sourceDir + "/" + path, destDir + "/" + path)

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