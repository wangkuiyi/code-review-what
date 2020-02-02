package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	betweenAngles = regexp.MustCompile(`<(.*)>`)
	betweenQuotes = regexp.MustCompile(`\"(.*)\"`)
)

type Comment struct {
	Body string `json: body`
}

type CommentList []Comment

func main() {
	url := "https://api.github.com/repos/sql-machine-learning/sqlflow/pulls/comments"

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.SetBasicAuth(os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_PASSWD"))

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
			fmt.Println(c.Body)
		}

		if _, ok := links["next"]; !ok {
			break
		}
		url = links["next"]
		fmt.Fprintf(os.Stderr, url)
	}
}
