package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	s "strings"

	"github.com/gocolly/colly"
)

type report struct {
	Languages map[string]int
	Repo []Repo
}

type Repo struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Language string `json:"language"`
	TotalStars string `json:"total_stars"`
	Issues string `json:"issues"`
	PR string `json:"pr"`
	URL string `json:"url"`
}

func main() {
	url := "https://github.com/trending/"
	c := colly.NewCollector()
	d := c.Clone()
	counter := 0

	var rep report

	repos := []Repo{}
	languages := &rep.Languages
	fmt.Println("languages are here:", languages)


	c.OnHTML("article.Box-row h1", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.ChildAttr("a", "href"))
		log.Println("visiting", link)
		d.Visit(link)
	})

	d.OnHTML("div.application-main", func(e *colly.HTMLElement) {
		repo := Repo{}

		name := e.ChildText("strong a")
		if name != "" {
			name = s.Replace(name, "-", " ", -1)
		} else {
			name = "Attribute missing"
		}
		repo.Name = name

		url := e.Request.URL.String()
		repo.URL = url
		
		description := e.ChildText("p.f4")
		descriptionClean := "Attribute missing"
		if description != "" {
			descriptionClean = s.Split(description, "\n")[0]
		} 
		repo.Description = descriptionClean

		language := e.ChildText("li.d-inline a span")
		languageClean := "Attribute missing"
		if language != "" {
			r, _ := regexp.Compile("[A-Za-z+#]+")
			languageClean = r.FindString(language)
		} 
		repo.Language = languageClean

		totalStars := e.ChildText("#repo-stars-counter-star")
		repo.TotalStars = totalStars
		
		issues := e.ChildText("#issues-repo-tab-count")
		repo.Issues = issues

		pr := e.ChildText("#pull-requests-repo-tab-count")
		repo.PR = pr

		repos = append(repos, repo)
	})

	d.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	d.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		counter += 1
		fmt.Println("Counter is:", counter)
		if counter == 25 {
			rep.Repo = repos
		
			js, err := json.MarshalIndent(repos, "", "    ")
			if err != nil {
				log.Fatal(err)
			}
			data := string(js)
			fmt.Println("data is:", data)

			// write to file
			if err := os.WriteFile("repos.json", js, 0664); err == nil {
				fmt.Println("Data written to file successfully")
			}
		}
	})

	c.Visit(url) 

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	enc.Encode(repos)
}