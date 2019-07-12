package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
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

func fetchPage(url string, isFile bool) (response *http.Response) {
	// Assume the connection wont be open longer than the timeout period set
	var httpClient = &http.Client{Timeout: 20 * time.Second}

	if isFile {
		httpClient.Timeout = 3 * time.Minute
	}

	response, err := httpClient.Get(url)
	checkErr(err)

	return
}

func parseFilename(url string) string {
	components := strings.Split(url, "/")
	return components[len(components)-1]
}

func downloadFile(filePath string, url string, waitGroup *sync.WaitGroup) {
	fmt.Println("Downloading:", url)
	response := fetchPage(url, true)
	defer response.Body.Close()

	outputFile, err := os.Create(filePath)
	checkErr(err)
	defer outputFile.Close()

	// io.Copy keeps the buffer to 32kb. Keeping memory efficient. It also dumps the data in the file itself.
	_, err = io.Copy(outputFile, response.Body)
	checkErr(err)

	defer func() {
		waitGroup.Done()
		fmt.Println("File Saved to:", filePath)
	}()
}

func getTunes(expansion string) (songs []string) {
	baseUrl := "https://ffxiv.tylian.net"
	url := baseUrl + "/" + expansion

	response := fetchPage(url, false)
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
					songs = append(songs, baseUrl+href)
				}
			}
		}
	}
}

func getExpansion() (expansion string) {
	expansions := map[int]string{
		0: "quit",
		1: "ffxiv",
		2: "ex1",
		3: "ex2",
		4: "ex3",
	}

	displayName := map[string]string{
		"quit":  "Quit",
		"ffxiv": "Final Fantasy Base Game",
		"ex1":   "Heavensward",
		"ex2":   "Stormblood",
		"ex3":   "Shadowbringers",
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter the number of the expansion you want downloaded")
	for key, val := range expansions {
		fmt.Printf("[%d] %s\n", key, displayName[val])
	}
	fmt.Print("> ")
	scanner.Scan()

	if result, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
		expansion = expansions[result]
		if expansion == "" {
			fmt.Println("Try Again")
			return getExpansion()
		} else if expansion == "quit" {
			os.Exit(0)
		}
	}

	return
}

func main() {
	var outputDir string

	flag.StringVar(&outputDir, "o", "./output", "Directory for songs to be downloaded to")
	flag.Parse()

	fmt.Println("This unofficial tool downloads songs from https://ffxiv.tylian.net\n")
	expansion := getExpansion()
	targetDir := strings.Join([]string{outputDir, expansion}, "/")

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Println("Creating", targetDir)
		os.MkdirAll(targetDir, os.ModeDir)
	}

	songs := getTunes(expansion)
	var waitGroup sync.WaitGroup

	for _, songUrl := range songs {
		waitGroup.Add(1)
		filePath := strings.Join([]string{targetDir, parseFilename(songUrl)}, "/")
		go downloadFile(filePath, songUrl, &waitGroup)
	}

	waitGroup.Wait()
	fmt.Println("Completed!")
}
