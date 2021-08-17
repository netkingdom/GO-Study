package scrapper

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type extractedJob struct {
	id       string
	title    string
	salary   string
	location string
	summary  string
}

// Scrape Indeed by term
func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/취업?q=" + term + "&limit=50"
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)
	fmt.Println(totalPages)

	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, c)
		//jobs = append(jobs, extractedJobs...)
	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}

	writeJobs(jobs)

	fmt.Println("Done, extracted ", len(jobs))
}

func getPage(page int, url string, mainC chan<- []extractedJob) {
	//slice로 묶어
	var jobs []extractedJob
	c := make(chan extractedJob)

	pageURL := url + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting ", pageURL)

	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCard := doc.Find(".tapItem")
	searchCard.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
		//jobs = append(jobs, job)
	})

	for i := 0; i < searchCard.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".jobTitle>span").Text())
	location := CleanString(card.Find(".companyLocation").Text())
	salary := CleanString(card.Find(".salary-snippet").Text())
	summary := CleanString(card.Find(".job-snippet").Text())

	c <- extractedJob{
		id:       id,
		title:    title,
		location: location,
		salary:   salary,
		summary:  summary,
	}
}

func getPages(url string) int {
	pages := 0

	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, selection *goquery.Selection) {
		pages = selection.Find("a").Length()
	})

	return pages
}

func writeJobs(jobs []extractedJob) {
	file, error := os.Create("jobs.csv") // 파일을만든다
	checkErr(error)

	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	file.Write(utf8bom)

	w := csv.NewWriter(file)
	defer w.Flush() // 함수가 끝날 떄 파일에 데이터를 입력

	headers := []string{"ID", "Title", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkErr(wErr)

	c := make(chan []string)

	for _, job := range jobs {
		go writeContent(job, c)
	}

	for i := 0; i < len(jobs); i++ {
		jobSlice := <-c
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}
func writeContent(content extractedJob, c chan<- []string) {
	c <- []string{
		content.id,
		content.title,
		content.location,
		content.salary,
		content.summary,
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		fmt.Println("Request failed with Status", res.StatusCode)
	}
}

// CleanString is Cleaninng String
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
