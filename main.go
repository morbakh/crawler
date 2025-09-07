package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

type Entry struct{ Header, Text, Url string }

var author_url string = "https://stihi.ru/avtor/dashkov"
var links_collection []string = nil

func main() {
	links_collection = collectLinks(author_url, 3500)
	total := len(links_collection)
	fmt.Println(total)

	for i, entry_url := range links_collection {
		entry_page := ParsePage(entry_url)
		entry := ReadEntry(entry_page, entry_url)
		SaveEntry("archive.txt", entry)
		if cent := i % 100; cent == 0 {
			fmt.Println(i)
		}
	}
}

func collectLinks(base_url string, max_pages int) (collection []string) {
	for i := 0; i <= max_pages; i += 50 {
		page_url := base_url + "&s=" + strconv.Itoa(i)
		doc := ParsePage(page_url)
		doc.Find(".maintext index ul li").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Find("a").Attr("href")
			collection = append(collection, "https://stihi.ru/"+link)
		})
	}

	return collection
}

func ReadEntry(doc *goquery.Document, entry_url string) (entry Entry) {
	entry.Header = doc.Find(".maintext index h1").Text()
	entry.Text = doc.Find(".maintext index .text").Text()
	entry.Url = entry_url

	return entry
}

func ParsePage(url string) (doc *goquery.Document) {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	utf8Reader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		log.Fatalf("ошибка преобразования кодировки: %v", err)
	}

	// Load the HTML document
	doc, err = goquery.NewDocumentFromReader(utf8Reader)

	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func SaveEntry(filename string, entry Entry) error {
	// Открываем файл для добавления (если не существует - создается)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "*****\n\n%s\n%s\n%s\n\n", entry.Header, entry.Text, entry.Url)
	if err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}

	return nil
}
