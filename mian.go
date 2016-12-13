package main

import (
	"fmt"
	"log"
	"os"

	"strings"

	"strconv"

	"github.com/spf13/cobra"
)

var file = ""

var rootCmd = &cobra.Command{
	Use:   "gond",
	Short: "go new definition",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		path, pos, err := splitPath(file)
		if err != nil {
			log.Fatalln(err)
		}
		def, err := findDefinition(os.Getenv("GOROOT"), os.Getenv("GOPATH"), path, pos)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(def)
	},
}

func splitPath(path string) (string, int, error) {
	ps := strings.LastIndex(path, "#")
	if ps < 0 {
		return "", 0, fmt.Errorf("there is no '#' in the path")
	}
	number := path[ps+1:]
	path = path[:ps]
	pos, err := strconv.Atoi(number)
	if err != nil {
		return "", 0, fmt.Errorf("line number not valid")
	}
	return path, pos, nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	// path: /path/to/src/file/filename.go#linenumber
	rootCmd.PersistentFlags().StringVarP(&file, "path", "p", "", "path of src file with line number")
	if err := rootCmd.Execute(); err != nil {
		log.Println("error:", err)
		os.Exit(-1)
	}
}
