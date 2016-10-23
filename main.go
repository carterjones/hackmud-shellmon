package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/fsnotify/fsnotify"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readShell(p string) {
	dat, err := ioutil.ReadFile(p)
	check(err)
	lines := strings.Split(string(dat), "\n")

	// TODO: extract the QR code here
	for _, l := range lines[1900:] {
		fmt.Println(l)
	}

	// return a QR code
}

func main() {
	path := flag.String("path", "shell.txt", "path to the shell.txt file")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					readShell(*path)
				}

				// TODO: take the qr code and translate the array to an image.Image with the 1's being black and the 0's being white

				// TODO: parse the image with https://github.com/kdar/goquirc

				// TODO: put the result into the clipboard
				result := "decoded result"
				log.Println(result)
				err := clipboard.WriteAll(result)
				if err != nil {
					log.Fatal(err)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(*path)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
