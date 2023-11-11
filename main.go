package main

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yakimka/goshort/storage"
)

func main() {
	storage, err := storage.NewSQLiteURLStorage("./urls.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	storage.Init()
	urlIdRegexp, err := regexp.Compile(`^[a-zA-Z0-9]{4,12}$`)
	if err != nil {
		log.Fatal(err)
	}
	registerHandlers(storage, urlIdRegexp, http.HandleFunc)

	log.Println("Listing for requests at http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

type HandlerRegistrator func(pattern string, handler func(http.ResponseWriter, *http.Request))

func registerHandlers(store storage.URLStorage, urlIdRegexp *regexp.Regexp, registrator HandlerRegistrator) {
	redirectHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.RawQuery)
		log.Println(r.URL.Query().Get("redirect"))
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}
		id, err := parseUrlId(r.URL.Path, urlIdRegexp)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		url, err := store.Get(id)
		log.Println(url)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		redirectParam := strings.ToLower(r.URL.Query().Get("redirect"))
		if redirectParam == "false" {
			io.WriteString(w, url)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
	}
	createHtmlFormHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html")
			io.WriteString(w, `<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Document</title>
			</head>
			<body>
				<form action="/api/v1/urls" method="POST">
					<input type="text" name="url" id="url">
					<input type="submit" value="submit">
				</form>
			</body>
			</html>`)
			return
		}
		if r.Method == http.MethodPost {
			url := r.FormValue("url")
			hashedUrl := hashURL(url)
			store.Set(hashedUrl, url)
			io.WriteString(w, "http://localhost:8000/"+hashedUrl)
			return
		}
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)

	}

	registrator("/", redirectHandler)
	registrator("/api/v1/urls", createHtmlFormHandler)
}

func parseUrlId(path string, idRegexp *regexp.Regexp) (string, error) {
	trimmedPath := strings.Trim(path, "/")
	if !idRegexp.MatchString(trimmedPath) {
		return "", errors.New("invalid URL ID")
	}
	return trimmedPath, nil
}

func hashURL(text string) string {
	const (
		// IEEE is by far and away the most common CRC-32 polynomial.
		// Used by ethernet (IEEE 802.3), v.42, fddi, gzip, zip, png, ...
		IEEE = 0xedb88320 // Castagnoli's polynomial, used in iSCSI.
	)

	crc32q := crc32.MakeTable(IEEE)
	checksum := crc32.Checksum([]byte(text), crc32q)
	return fmt.Sprintf("%08x", checksum)
}
