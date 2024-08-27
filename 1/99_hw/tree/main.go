package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main_.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	fmt.Println(printFiles)
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	tree := ""
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	var last int
	for i, file := range files {
		if file.IsDir() {
			last = i
		}
	}
	for i, file := range files {
		if file.IsDir() {
			flag := i == len(files)-1 || (i == last && !printFiles)
			add, err := recursive(path+string(os.PathSeparator)+file.Name(), file.Name(), printFiles, "", flag)
			if err != nil {
				return err
			}
			if flag {
				tree += "└───" + add
			} else {
				tree += "├───" + add
			}
		} else if printFiles {
			info, err := file.Info()
			if err != nil {
				return err
			}
			var size string
			if info.Size() == 0 {
				size = "empty"
			} else {
				size = fmt.Sprintf("%db", info.Size())
			}
			if i == len(files)-1 {
				tree += "└───" + file.Name() + fmt.Sprintf(" (%v)", size)
			} else {
				tree += "├───" + file.Name() + fmt.Sprintf(" (%v)", size)
			}
		}
	}
	_, err = fmt.Fprintln(out, strings.TrimRight(tree, "\n"))
	return err
}

func recursive(path string, cur string, printFiles bool, ot string, flag bool) (string, error) {
	tree := cur + "\n"
	if flag {
		ot += "\t"
	} else {
		ot += "│\t"
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	for i, file := range files {
		if file.IsDir() {
			var add string
			flag = i == len(files)-1
			add, err = recursive(path+string(os.PathSeparator)+file.Name(), file.Name(), printFiles, ot, flag)
			if err != nil {
				return "", err
			}
			if flag {
				tree += ot + "└───" + add
			} else {
				tree += ot + "├───" + add
			}
		} else if printFiles {
			info, err := file.Info()
			if err != nil {
				return "", err
			}
			var size string
			if info.Size() == 0 {
				size = "empty"
			} else {
				size = fmt.Sprintf("%db", info.Size())
			}
			flag = i == len(files)-1
			if flag {
				tree += ot + "└───" + file.Name() + fmt.Sprintf(" (%v)", size) + "\n"
			} else {
				tree += ot + "├───" + file.Name() + fmt.Sprintf(" (%v)", size) + "\n"
			}
		}
	}
	return tree, nil
}
