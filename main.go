package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/auburnsummer/glamour"
)

const DEFAULT_ERROR = "I could not find that page."
const THEME = "dracula"

// try to download a file. the channel will resolve with either the contents, or an empty byte slice.
func DownloadFileToChannel(url string, out chan<- []byte) {
	empty := []byte{}
	resp, err := http.Get(url)
	if err != nil {
		out <- empty
		return
	}

	if resp.StatusCode != 200 {
		out <- empty
		return
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		out <- empty
		return
	}

	glamouredString, err := glamour.RenderBytes(content, THEME)
	if err != nil {
		out <- empty
		return
	}

	out <- glamouredString
}

func GetGlamourTldrPage(page string) (out []byte) {
	// platforms TLDR pages supports.
	platforms := [4]string{"common", "linux", "osx", "windows"}

	// corresponding URLs they must be hosted on if they exist.
	urls := platforms // this is a copy
	for i, plat := range platforms {
		urls[i] = fmt.Sprintf("https://raw.githubusercontent.com/tldr-pages/tldr/main/pages/%s/%s.md", plat, page)
	}

	var chans [4]chan []byte
	// corresponding channels we'll get the results back.
	for i := range platforms {
		newBox := make(chan []byte, 1)
		chans[i] = newBox
	}

	for i, plat := range platforms {
		url := fmt.Sprintf("https://raw.githubusercontent.com/tldr-pages/tldr/main/pages/%s/%s.md", plat, page)
		go DownloadFileToChannel(url, chans[i])
	}

	for i := range platforms {
		result := <-chans[i]
		if len(result) > 0 {
			return result
		}
	}
	return []byte{}
}

func handler(w http.ResponseWriter, req *http.Request) {
	url := req.URL
	path := strings.TrimPrefix(url.Path, "/")

	out := GetGlamourTldrPage(path)

	w.Write(out)
}

func main() {
	// Hello world, the web server

	http.HandleFunc("/", handler)
	log.Println("Listing for requests at http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}