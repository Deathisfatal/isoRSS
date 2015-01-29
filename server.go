package main

import (
	"fmt"
	. "github.com/gorilla/feeds"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

var (
	now = time.Now()

	feed = &Feed{
		Title:       "Linux ISO BitTorrent feed",
		Link:        &Link{Href: "http://github.com/Deathisfatal/isoRSS"},
		Description: "RSS for various Linux distribution ISOs",
		Author:      &Author{"Isaac True", "isaac.true@gmail.com"},
		Created:     now,
		Copyright:   "",
	}

	debianBaseURL = "http://cdimage.debian.org/debian-cd/current/"
)

type ISO struct {
	distro   string
	url      string
	filename string
}

func parseHrefsFromDebianPage(page string, arch string, _type string) []string {
	r, _ := regexp.Compile("href=\"([^\"]*.torrent)\"")
	hrefs := r.FindAllString(page, -1)
	for i, element := range hrefs {
		hrefs[i] = debianBaseURL + arch + "/" + _type + element[6:len(element)-1]
	}
	return hrefs
}

func fetchDebianPage(arch string, _type string) string {
	baseurl := "http://cdimage.debian.org/debian-cd/current/"
	resp, _ := http.Get(baseurl + arch + "/" + _type)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func fetchDebianISOs() []ISO {
	isos := make([]ISO, 0)
	cd := "bt-cd/"
	archs := []string{"amd64", "i386", "armel", "mips", "armhf"}

	for _, arch := range archs {
		hrefs := parseHrefsFromDebianPage(fetchDebianPage(arch, cd), arch, cd)
		for _, href := range hrefs {
			isos = append(isos, ISO{
				distro:   "Debian",
				url:      href,
				filename: href,
			})
		}
	}

	return isos
}

func fetchISOs() []ISO {
	isos := make([]ISO, 0)
	isos = append(isos, fetchDebianISOs()...)
	return isos
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8888", nil)
}

func makeItemFromISO(iso ISO) Item {
	item := Item{
		Title:       iso.filename,
		Link:        &Link{Href: iso.url},
		Description: iso.distro,
		Id:          "0",
		Updated:     now,
		Created:     now,
		Author:      &Author{iso.distro, ""},
	}
	return item
}

func handler(w http.ResponseWriter, r *http.Request) {
	isos := fetchISOs()
	items := make([]Item, len(isos))
	for i, element := range isos {
		items[i] = makeItemFromISO(element)
	}
	itemPointers := make([]*Item, len(items))
	for i, element := range items {
		itemPointers[i] = &element
	}
	feed.Items = itemPointers
	rss, err := feed.ToRss()
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "%s", rss)
}
