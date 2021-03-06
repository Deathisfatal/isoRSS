package main

import (
	"fmt"
	. "github.com/gorilla/feeds"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
    "io"
    "encoding/hex"
    "crypto/md5"
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
    date        time.Time
    guid string
}

type debianISO struct {
    href string
    date time.Time
}

func parseISOFromDebianPage(page string, arch string, _type string) []debianISO {
	r, _ := regexp.Compile("href=\"([^\"]*.torrent)\"")
	hrefs := r.FindAllString(page, -1)
    iso := make([]debianISO, len(hrefs))
	for i, element := range hrefs {
		hrefs[i] = debianBaseURL + arch + "/" + _type + element[6:len(element)-1]
        iso[i] = debianISO{
            href: hrefs[i],
            date: now, //todo
        }
	}
	return iso
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
	archs := []string{
        "amd64",
        "i386",
       /* "armel",*/
       /* "mips",*/
       /* "armhf",*/
       /* "ia64",*/
       /* "kfreebsd-i386",*/
       /* "kfreebsd-amd64",*/
       /* "mipsel",*/
       /* "powerpc",*/
       /* "sparc",*/
       /* "s390", */
       /* "s390x", */
       /* "source",*/
        "multi-arch",
    }
	for _, arch := range archs {
		debianiso := parseISOFromDebianPage(fetchDebianPage(arch, cd), arch, cd)
		for _, iso := range debianiso {
            md5Gen := md5.New()
            io.WriteString(md5Gen, iso.href)
            hash := hex.EncodeToString(md5Gen.Sum(nil))
			isos = append(isos, ISO{
				distro:   "Debian",
				url:      iso.href,
				filename: iso.href,
                date:     iso.date,
                guid:     hash,
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
	http.ListenAndServe(":10000", nil)
}

func makeItemFromISO(iso ISO) Item {
	item := Item{
		Title:       iso.filename,
		Link:        &Link{Href: iso.url},
		Description: iso.distro,
		Id:          iso.guid,
		Updated:     iso.date,
		Created:     iso.date,
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
	for i, _ := range items {
		itemPointers[i] = &items[i]
	}
	feed.Items = itemPointers
	rss, err := feed.ToRss()
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "%s", rss)
}
