package main

import (
	"fmt"
	"github.com/labstack/echo"
	"jobScrapper/scrapper"
	"os"
	"strings"
)

const fileName string = "jobs.csv"

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	// Start server
	e.Logger.Fatal(e.Start(":1323"))



:Q	fmt.Println("done")
	fmt.Println("checkoute")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(fileName)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment(fileName, fileName)

}

// Handler
func handleHome(c echo.Context) error {
	return c.File("home.html")
}
