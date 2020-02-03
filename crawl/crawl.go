package main

import (
	"encoding/json"
	"flag"
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
	repo := flag.String("repo", "sql-machine-learning/sqlflow", "GitHub repo name in the format of org/repo")
	user := flag.String("user", "", "GitHub account username.  Could be empty if don't authorize")
	passwd := flag.String("passwd", "", "GitHub account password.  Could be empty if don't authorize")
	flag.Parse()

	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/comments", *repo)

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
			fmt.Println(c.Body)
		}

		if _, ok := links["next"]; !ok {
			break
		}
		url = links["next"]
		fmt.Fprintf(os.Stderr, "%s\n", url)
	}
}
