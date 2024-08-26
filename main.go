package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

// Job represents the structure of job data to be scraped.
type Job struct {
	Jobtitle string `json:"title"`
	Name     string `json:"name"`
	City     string `json:"city"`
	Salary   string `json:salary`
}

// Scraper handles the scraping logic.
type Scraper struct {
	URL      string
	JobsChan chan Job
	WG       sync.WaitGroup
}

// NewScraper creates a new Scraper instance.
func NewScraper(url string) *Scraper {
	return &Scraper{
		URL:      url,
		JobsChan: make(chan Job),
	}
}

// Start initiates the scraping process.
func (s *Scraper) Start() {
	s.WG.Add(1)
	go s.scrape()

	// Collect results from the channel
	go func() {
		s.WG.Wait()
		close(s.JobsChan)
	}()
}

// scrape performs the actual web scraping.
func (s *Scraper) scrape() {
	defer s.WG.Done()

	c := colly.NewCollector(
		colly.AllowedDomains("internshala.com"), // Correct domain
	)

	c.OnHTML("div.internship_meta", func(element *colly.HTMLElement) {
		job := Job{
			Jobtitle: element.ChildText(".job-internship-name"),
			Name:     element.ChildText(".company-name"),
			City:     element.ChildText(".locations"),
			Salary:   element.ChildText(".desktop"),
		}

		s.JobsChan <- job
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL)
		// Add a random delay to avoid getting blocked
		delay := time.Duration(rand.Intn(5)+1) * time.Second
		fmt.Printf("Sleeping for %v\n", delay)
		time.Sleep(delay)
	})

	if err := c.Visit(s.URL); err != nil {
		log.Printf("Error visiting %s: %v", s.URL, err)
	}
}

// writeJSON writes the scraped job data to a JSON file.
func writeJSON(data []Job) {
	f, err := json.MarshalIndent(data, "", "  ") // Indentation for readability
	if err != nil {
		log.Fatal(err)
		return
	}

	if err := os.WriteFile("job.json", f, 0644); err != nil {
		log.Fatal(err)
	}
}

func main() {
	url := "https://internshala.com/jobs/"
	scraper := NewScraper(url)

	scraper.Start()

	var jobs []Job
	for job := range scraper.JobsChan {
		jobs = append(jobs, job)
	}

	writeJSON(jobs)
}
