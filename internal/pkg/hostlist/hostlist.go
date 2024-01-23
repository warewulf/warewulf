package hostlist

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func toInt(i string) int {
	ret, _ := strconv.Atoi(i)
	return ret
}

func isDigit(s string) bool {
	r := []rune(s)

	for i := 0; i < len(r); i++ {
		if !unicode.IsDigit(r[i]) {
			return false
		}
	}

	return true
}

func expand_iterate(list []string) ([]string, int) {
	var ret []string
	var count int

	for _, i := range list {

		bracketIndex1 := strings.Index(i, "[")
		bracketIndex2 := strings.Index(i, "]")

		if bracketIndex1 >= 0 && bracketIndex1 < bracketIndex2 {
			prefix := i[:bracketIndex1]
			suffix := i[bracketIndex2+1:]
			ranges := strings.Split(i[bracketIndex1+1:bracketIndex2], ",")
			count++

			for _, r := range ranges {
				iterate := strings.Split(r, "-")

				if len(iterate) == 1 {
					if !isDigit(iterate[0]) {
						return ret, 0
					} else {
						ret = append(ret, prefix+iterate[0]+suffix)
					}
				} else if len(iterate) == 2 {
					if !isDigit(iterate[0]) || !isDigit(iterate[1]) {
						return ret, 0
					} else {
						sigfigures := len(iterate[0])
						for i := toInt(iterate[0]); i <= toInt(iterate[1]); i++ {
							ret = append(ret, fmt.Sprintf(`%s%.`+fmt.Sprintf("%d", sigfigures)+`d%s`, prefix, i, suffix))
						}
					}
				}
			}
		}
	}
	return ret, count
}

func Expand(list []string) []string {
	ret := list

	for {
		loop, count := expand_iterate(ret)

		if count == 0 {
			break
		}
		ret = loop
	}

	return ret
}
