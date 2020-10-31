
package util

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func FileCopyReplace(sourceFile string, destFile string, replace map[string]string) error{

    sourceFD, err := os.Open(sourceFile)
    if err != nil {
        return err
    }

    destFD, err := os.Create(destFile)
    if err != nil {
        return err
    }
    defer sourceFD.Close()
    defer destFD.Close()

    scanner := bufio.NewScanner(sourceFD)
    w := bufio.NewWriter(destFD)

    for scanner.Scan() {
        newLine := scanner.Text()
        for k, v := range replace {
            replaceString := fmt.Sprintf("@{%s}", strings.ToUpper(k))
            newLine = strings.Replace(newLine, replaceString, v, -1)

        }
        w.WriteString(newLine + "\n")

    }

    return w.Flush()
}

