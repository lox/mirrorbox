package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/abh/geoip"
)

const (
	BIND_ADDR   = ":80"
	MIRRORS_URL = "http://mirrors.ubuntu.com/%s.txt"
	STATUS_URL  = "https://launchpad.net/ubuntu/+archivemirrors"
	GEOIP_DB    = "/usr/share/GeoIP/GeoIP.dat"
)

var (
	ALLOWED_STATES = []string{
		"distromirrorstatusONEHOURBEHIND",
		"distromirrorstatusTWOHOURSBEHIND",
		"distromirrorstatusUP",
	}
	status    map[string]bool
	lastFetch int64
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

// gets an array of mirrors for a given country
func mirrors(country string) ([]string, error) {
	url := fmt.Sprintf(MIRRORS_URL, strings.ToUpper(country))

	log.Printf("fetching %s", url)
	response, err := http.Get(url)
	if err != nil {
		return []string{}, err
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []string{}, err
	}

	return strings.Split(string(contents), "\n"), nil
}

func countryCode(r *http.Request) (string, error) {
	remoteHostParts := strings.SplitN(r.RemoteAddr, ":", 2)

	file := GEOIP_DB
	if dbenv := os.Getenv("GEOIP_DB"); dbenv != "" {
		file = dbenv
	}

	log.Printf("using %s for GeoIP database, looking up %s",
		file, remoteHostParts[0])

	gi, err := geoip.Open(file)
	if err != nil {
		return "", err
	}

	if gi != nil {
		country, _ := gi.GetCountry(remoteHostParts[0])
		return country, nil
	}

	return "", nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	country, _ := countryCode(r)
	if country == "" {
		log.Printf("failed to lookup %s, defaulting to US", r.RemoteAddr)
		country = "US"
	}

	mirrors, err := mirrors(country)
	if err != nil {
		http.Error(w, "Failed to fetch mirrors", http.StatusInternalServerError)
		return
	}

	if time.Now().After(time.Unix(lastFetch, 0).Add(time.Minute * 60)) {
		status, err = mirrorStatus()
		if err != nil {
			http.Error(w, "Failed to fetch mirror status", http.StatusInternalServerError)
			return
		}
		lastFetch = time.Now().Unix()
	}

	// only output legit mirrors
	for _, mirror := range mirrors {
		if s, ok := status[mirror]; ok && s {
			w.Write([]byte(mirror + "\n"))
		}
	}
}

func main() {
	bindaddr := BIND_ADDR

	// allow bind address to be overriden
	if bindenv := os.Getenv("BIND_ADDR"); bindenv != "" {
		bindaddr = bindenv
	}

	http.HandleFunc("/mirrors.txt", handler)
	log.Printf("listening on %s", bindaddr)
	if err := http.ListenAndServe(bindaddr, nil); err != nil {
		log.Fatalf("listen failed: %s", err.Error())
	}
}
