package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func getHref(token html.Token) (ok bool, href string) {
	// Loop through token attributes and return the value
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			href = attr.Val
			ok = true
		}
	}
	return
}

func fetchPage(url string) (response *http.Response) {
	var httpClient = &http.Client{Timeout: 10 * time.Second}
	response, err := httpClient.Get(url)
	checkErr(err)

	return
}

func downloadFile(filePath string, url string) error {
	response := fetchPage(url)
	defer response.Body.Close()

	outputFile, err := os.Create(filePath)
	checkErr(err)
	defer outputFile.Close()

	// io.Copy keeps the buffer to 32kb. Keeping memory efficient.
	_, err = io.Copy(outputFile, response.Body)
	return err
}

func getTunes() (songs []string) {
	baseUrl := "https://ffxiv.tylian.net"
	expansion := "ex3"

	url := baseUrl + "/" + expansion

	response := fetchPage(url)
	defer response.Body.Close()

	tokenizer := html.NewTokenizer(response.Body)

	for {
		tag := tokenizer.Next()

		switch {
		case tag == html.ErrorToken:
			return
		case tag == html.StartTagToken:
			token := tokenizer.Token()

			if token.Data == "a" {
				ok, href := getHref(token)
				if !ok {
					continue
				}

				isSong := strings.HasSuffix(href, "mp3")

				if isSong {
					// fmt.Printf("Downloading: %s%s\n", baseUrl, href)
					songs = append(songs, baseUrl+href)
				}
			}
		}
	}
}

func main() {
	for _, song := range getTunes() {
		fmt.Println(song)
	}
}
