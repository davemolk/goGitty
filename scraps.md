package main

import (
	"fmt"
	"log"
	s "strings"

	"github.com/gocolly/colly"
)

type Repo struct {
	Name string
	Description string
	Language string
	TotalStars string
	Issues string
	PR string
	URL string
}

func main() {
	url := "https://github.com/trending/"
	c := colly.NewCollector()
	d := c.Clone()

	repos := make([]Repo, 0)

	c.OnHTML("article.Box-row h1", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.ChildAttr("a", "href"))
		log.Println("visiting", link)
		d.Visit(link)
	})

	d.OnHTML("div.application-main", func(e *colly.HTMLElement) {
		repo := Repo{}

		name := e.ChildText("strong a")
		nameClean := s.Replace(name, "-", " ", -1)
		fmt.Println("name:", nameClean)
		repo.Name = name

		url := e.Request.AbsoluteURL(e.ChildAttr("a", "href"))
		// fmt.Println("url:", url)
		repo.URL = url
		
		description := e.ChildText("p.f4")
		descriptionClean := s.Split(description, "\n")
		fmt.Println("description:", descriptionClean[0])
		// fmt.Println("description:", description)
		repo.Description = description

		language := e.ChildText("li.d-inline a span")
		// fmt.Println("language:", language)
		repo.Language = language

		totalStars := e.ChildText("#repo-stars-counter-star")
		// fmt.Println("totalStars:", totalStars)
		repo.TotalStars = totalStars
		
		issues := e.ChildText("#issues-repo-tab-count")
		// fmt.Println("issues:", issues)
		repo.Issues = issues

		pr := e.ChildText("#pull-requests-repo-tab-count")
		// fmt.Println("pr:", pr)
		repo.PR = pr

		repos = append(repos, repo)
	})

	d.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// d.OnScraped(func(r *colly.Response) {
	// 	fmt.Println("Finished", r.Request.URL)
	// 	js, err := json.MarshalIndent(repos, "", "    ")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println("Writing data to file")
	// 	if err := os.WriteFile("repos.json", js, 0664); err == nil {
	// 		fmt.Println("Data written to file successfully")
	// 	}
	// })

	c.Visit(url)
}