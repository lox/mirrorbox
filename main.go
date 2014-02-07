package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	MIRRORS_URL = "http://mirrors.ubuntu.com/mirrors.txt"
	STATUS_URL  = "https://launchpad.net/ubuntu/+archivemirrors"
)

var (
	ALLOWED_STATES = []string{
		"distromirrorstatusONEHOURBEHIND",
		"distromirrorstatusTWOHOURSBEHIND",
		"distromirrorstatusUP",
	}
)

func fetchMirrorStatusHtml() (string, error) {
	log.Printf("fetching %s", STATUS_URL)
	response, err := http.Get(STATUS_URL)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func mirrorStatus() (map[string]bool, error) {
	status := map[string]bool{}
	html, err := fetchMirrorStatusHtml()
	if err != nil {
		return status, err
	}

	trRe := regexp.MustCompile("(?si)<tr>(.+?)</tr>")
	tagRe := regexp.MustCompile("<(a|span) (class|href)=\"(.+?)\">.+?</(a|span)>")

	for _, tr := range trRe.FindAllStringSubmatch(html, -1) {
		matches := tagRe.FindAllStringSubmatch(tr[1], -1)
		if len(matches) > 0 {
			state := false
			for _, v := range ALLOWED_STATES {
				if v == matches[len(matches)-1][3] {
					state = true
					break
				}
			}

			status[matches[1][3]] = state
		}
	}

	return status, nil
}

func mirrors() ([]string, error) {
	log.Printf("fetching %s", MIRRORS_URL)
	response, err := http.Get(MIRRORS_URL)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []string{}, err
	}

	return strings.Split(string(contents), "\n"), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	mirrors, err := mirrors()
	if err != nil {
		log.Fatalf("Failed to get mirrors: %s", err.Error())
	}

	status, err := mirrorStatus()
	if err != nil {
		log.Fatalf("Failed to get mirrors status: %s", err.Error())
	}

	// only output legit mirrors
	for _, mirror := range mirrors {
		if s, ok := status[mirror]; ok && s {
			w.Write([]byte(mirror + "\n"))
		}
	}
}

func main() {
	http.HandleFunc("/mirrors.txt", handler)

	log.Printf("Listening on :8900")
	http.ListenAndServe(":8900", nil)
}
