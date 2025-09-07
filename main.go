package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

type Entry struct{ Header, Text, Url string }

const stihiBaseHost = "https://stihi.ru/"

func main() {
	var (
		linksCollection []string
		authorUrl       = stihiBaseHost + "avtor/dashkov"
	)

	linksCollection, err := collectLinksConc(authorUrl, 3500)
	if err != nil {
		fmt.Printf("ошибка сбора ссылок %v", err)
		return
	}

	total := len(linksCollection)
	fmt.Println(total)

	for i, entryUrl := range linksCollection {
		entryPage, err := parsePage(entryUrl)
		if err != nil {
			fmt.Printf("ошибка парсинга страницы %v", err)
			continue
		}

		err = saveEntry("archive.txt", readEntry(entryPage, entryUrl))
		if err != nil {
			fmt.Printf("не удалось сохранить запись %s: %v", entryUrl, err)
		}

		if i%100 == 0 {
			fmt.Println(i)
		}
	}
}

func collectLinksConc(baseURL string, maxPages int) ([]string, error) {
	var links []string

	for i := 0; i <= maxPages; i += 50 {
		pageURL := baseURL + "&s=" + strconv.Itoa(i)

		doc, err := parsePage(pageURL)
		if err != nil {
			return nil, fmt.Errorf("пропускаю страницу %s из-за ошибки: %v", pageURL, err)
		}

		doc.Find(".maintext index ul li").Each(func(i int, s *goquery.Selection) {
			if link, exists := s.Find("a").Attr("href"); exists {
				links = append(links, stihiBaseHost+link)
			}
		})
	}

	return links, nil
}

func readEntry(doc *goquery.Document, entryUrl string) (entry Entry) {
	entry.Header = doc.Find(".maintext index h1").Text()
	entry.Text = doc.Find(".maintext index .text").Text()
	entry.Url = entryUrl

	return entry
}

func parsePage(url string) (*goquery.Document, error) {
	time.Sleep(10 * time.Millisecond)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("статус ответа: %d %s", res.StatusCode, res.Status)
	}

	utf8Reader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("ошибка преобразования кодировки: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга страницы: %v", err)
	}
	return doc, nil
}

func saveEntry(filename string, entry Entry) error {
	// Открываем файл для добавления (если не существует - создается)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "*****\n\n%s\n%s\n%s\n\n", entry.Header, entry.Text, entry.Url)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	return nil
}
