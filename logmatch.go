/*
history:
20.1114 v1

GoFmt GoBuild GoRelease
GoRun '^Request starting ' '^Request finished in [^ ]+ 500 ' logger*.txt
*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	LogmatchDir string
)

func log(msg string, args ...interface{}) {
	const Beat = time.Duration(24) * time.Hour / 1000
	tzBiel := time.FixedZone("Biel", 60*60)
	t := time.Now().In(tzBiel)
	ty := t.Sub(time.Date(t.Year(), 1, 1, 0, 0, 0, 0, tzBiel))
	td := t.Sub(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tzBiel))
	ts := fmt.Sprintf(
		"%d/%d@%d",
		t.Year()%1000,
		int(ty/(time.Duration(24)*time.Hour))+1,
		int(td/Beat),
	)
	fmt.Fprintf(os.Stderr, ts+" "+msg+"\n", args...)
}

func init() {
	var err error

	LogmatchDir = os.Getenv("LogmatchDir")
	if LogmatchDir == "" {
		LogmatchDir = "."
	}

	LogmatchDir, err = filepath.Abs(LogmatchDir)
	if err != nil {
		log("Cannot get absolute path for dir `%s`: %v", LogmatchDir, err)
		os.Exit(1)
	}

	err = os.MkdirAll(LogmatchDir, 0700)
	if err != nil {
		log("Cannot make dir `%s`: %v", LogmatchDir, err)
		os.Exit(1)
	}
}

func main() {
	var err error

	if len(os.Args) < 3 {
		log("usage: logmatch starting.line.regexp finishing.line.regexp file.path...")
		os.Exit(1)
	}

	startRe, err := regexp.Compile(os.Args[1])
	if err != nil {
		log("Regexp `%s` invalid: %v", os.Args[1], err)
		os.Exit(1)
	}
	finishRe, err := regexp.Compile(os.Args[2])
	if err != nil {
		log("Regexp `%s` invalid: %v", os.Args[2], err)
		os.Exit(1)
	}

	for _, logpath := range os.Args[3:] {
		file, err := os.OpenFile(logpath, os.O_RDONLY, 0600)
		if err != nil {
			log("Cannot open file `%s`: %v", logpath, err)
			os.Exit(1)
		}
		filereader := bufio.NewReader(file)

		var ll []string
		var mi int

		for {
			l, err := filereader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log("ReadString: %v", err)
				}
				break
			}

			if startRe.MatchString(l) {
				ll = []string{l}
			} else {
				ll = append(ll, l)
				if finishRe.MatchString(l) {
					mi++
					reportpath := path.Join(
						LogmatchDir,
						fmt.Sprintf("%s.logmatch.%d.text", path.Base(logpath), mi))
					err = ioutil.WriteFile(
						reportpath,
						[]byte(strings.Join(ll, "")),
						0600)
					if err != nil {
						log("Could not write file `%s`: %v", reportpath, err)
						os.Exit(1)
					}
				}
			}
		}

		file.Close()
	}

}
