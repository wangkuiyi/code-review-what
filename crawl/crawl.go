package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	betweenAngles = regexp.MustCompile(`<(.*)>`)
	betweenQuotes = regexp.MustCompile(`\"(.*)\"`)
)

type User struct {
	Login string `json: login`
}

type Comment struct {
	Body       string `json: body`
	User       `json: user`
	CreateTime string `json:created_at`
	UpdateTime string `json:updated_at`
}

type CommentList []Comment

func main() {
	user := flag.String("user", "", "GitHub account username.  Could be empty if don't authorize")
	passwd := flag.String("passwd", "", "GitHub account password.  Could be empty if don't authorize")
	flag.Parse()

	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/comments", flag.Arg(0))
	log.Printf("Crawling %s", url)

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(*user) > 0 && len(*passwd) > 0 {
			req.SetBasicAuth(*user, *passwd)
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		headerLink := resp.Header["Link"][0]
		links := map[string]string{}
		for _, link := range strings.Split(headerLink, ",") {
			ss := strings.Split(link, ";")
			l := betweenAngles.FindStringSubmatch(ss[0])[1]
			t := betweenQuotes.FindStringSubmatch(ss[1])[1]
			links[t] = l
		}

		var cl CommentList
		err = json.NewDecoder(resp.Body).Decode(&cl)
		if err != nil {
			log.Fatal(err)
		}
		for _, c := range cl {
			comment := strings.Replace(
				strings.Replace(c.Body, "\r\n", " ", -1), // \r\n is in-comment line break.
				",", " ", -1)                             // CSV use commad to separate values.
			fmt.Printf("%s,%s,%s,%s\n", c.User.Login, c.CreateTime, c.UpdateTime, comment)
		}

		if _, ok := links["next"]; !ok {
			break
		}
		url = links["next"]
		log.Printf("%s\n", url)
	}
}
