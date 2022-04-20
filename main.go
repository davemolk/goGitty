package main

import (
	"encoding/json"
	"log"
	"net/smtp"
	"os"
	"regexp"
	s "strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

type report struct {
	ScrapedAt string
	Languages map[string]int
	Repos map[string]Repo
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

	// send email if alarm is triggered
	alarm := false

	scrapedAt := time.Now().Format("2006-01-02")
	repos := map[string]Repo{}
	languages := map[string]int{}
	rep := report{
		ScrapedAt: scrapedAt,
		Languages: languages,
		Repos: repos,
	}

	c.OnHTML("article.Box-row h1", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.ChildAttr("a", "href"))
		log.Println("visiting", link)
		d.Visit(link)
	})

	d.OnHTML("div.application-main", func(e *colly.HTMLElement) {
		repo := Repo{}
		errorWatcher := 0

		name := e.ChildText("strong a")
		if name != "" {
			name = s.Replace(name, "-", " ", -1)
		} else {
			name = "Attribute missing"
			errorWatcher += 1
		}
		repo.Name = name

		url := e.Request.URL.String()
		repo.URL = url
		
		description := e.ChildText("p.f4")
		descriptionClean := "Attribute missing"
		if description != "" {
			descriptionClean = s.Split(description, "\n")[0]
		} else {
			errorWatcher += 1
		}
		repo.Description = descriptionClean

		language := e.ChildText("li.d-inline a span")
		languageClean := "Attribute missing"
		if language != "" {
			r, _ := regexp.Compile("[A-Za-z+#]+")
			languageClean = r.FindString(language)
		} else {
			errorWatcher += 1
		}
		repo.Language = languageClean
		if languageClean != "Attribute missing" {
			languages[languageClean]++
		} 

		totalStars := e.ChildText("#repo-stars-counter-star")
		repo.TotalStars = totalStars
		
		issues := e.ChildText("#issues-repo-tab-count")
		repo.Issues = issues

		pr := e.ChildText("#pull-requests-repo-tab-count")
		repo.PR = pr

		repos[repo.Name] = repo

		// set threshold
		if errorWatcher > 2 {
			alarm = true
		}
	})

	d.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit(url) 

	rep.Languages = languages
	rep.Repos = repos


	js, err := json.MarshalIndent(rep, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	deets := string(js)
	log.Println("do something with data", deets)

	if alarm {
		message := []byte("possible issue with gitty, please check")
		email(message)
	}

} 

func email(message []byte) {
    err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	from := os.Getenv("FROM")
	password := os.Getenv("PASSWORD")

	to := []string{
		os.Getenv("TO"),
	}
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	body := message

	auth := smtp.PlainAuth("", from, password, smtpHost)

	addr := s.Join([]string{smtpHost, smtpPort}, ":")

	emailErr := smtp.SendMail(addr, auth, from, to, body)
	if emailErr != nil {
		log.Println(emailErr)
		return
	}
	log.Printf("Email successfully sent to %s", to[0])
}