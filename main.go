package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

type Entry struct{ Header, Text string }

func ParseEntry(url string) (entry Entry) {
	// Request the HTML page.
	res, err := http.Get("https://stihi.ru/2025/08/01/5781")
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
	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	entry.Header = doc.Find(".maintext index h1").Text()
	entry.Text = doc.Find(".maintext index .text").Text()

	return entry
}

func SaveEntry(filename string, entry Entry) error {
	// Открываем файл для добавления (если не существует - создается)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%s\n %s\n *** \n\n", entry.Header, entry.Text)
	if err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}

	return nil
}

func main() {
	entry := ParseEntry("https://stihi.ru/2025/08/01/5781")
	SaveEntry("archive.txt", entry)
}
