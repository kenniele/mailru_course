package main

import (
	"bufio"
	"fmt"
	"github.com/mailru/easyjson"
	"io"
	"os"
	"strings"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	/*
		!!! !!! !!!
		обратите внимание - в задании обязательно нужен отчет
		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировали их
		так же обратите внимание на команду с параметром -http
		перечитайте еще раз задание
		!!! !!! !!!
	*/
	//SlowSearch(out)
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	seenBrowsers := make(map[string]struct{}, 115)
	builder := strings.Builder{}
	builder.Grow(115)
	i := 0
	for scanner.Scan() {
		user := User{}
		err := easyjson.Unmarshal(scanner.Bytes(), &user)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		browsers := user.Browsers

		for _, browser := range browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				seenBrowsers[browser] = struct{}{}
			}
		}

		for _, browser := range browsers {
			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				seenBrowsers[browser] = struct{}{}
			}
		}

		if !(isAndroid && isMSIE) {
			i++
			continue
		}

		user.Email = strings.Replace(user.Email, "@", " [at] ", -1)
		builder.WriteString(fmt.Sprintf("[%v] %v <%v>\n", i, user.Name, user.Email))
		i++
	}

	fmt.Fprintln(out, "found users:\n"+builder.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
