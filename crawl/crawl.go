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

type User struct {
	Login string `json: login`
}

type Comment struct {
	Body string `json: body`
	User User   `json: user`
}

type Pull struct {
	Number int    `json: number,omitempty`
	User   User   `json: user,omitempty`
	Title  string `json: title,omitempty`
	Body   string `json: body,omitempty`
}

func main() {
	user := flag.String("user", os.Getenv("GITHUB_USER"), "GitHub username.")
	passwd := flag.String("passwd", os.Getenv("GITHUB_PASSWD"), "GitHub passwd.")
	pulls := flag.String("pulls", "pulls.csv", "Output CSV filename for pull requests.")
	comments := flag.String("comments", "comments.csv", "Output CSV filename for code review comments.")
	flag.Parse()

	f, e := os.Create(*pulls)
	if e != nil {
		log.Fatal(e)
	}
	defer f.Close()

	ff, e := os.Create(*comments)
	if e != nil {
		log.Fatal(e)
	}
	defer ff.Close()

	crawlPulls(flag.Arg(0), *user, *passwd, f, ff)
}

func crawlPulls(repo, user, passwd string, pulls, comments *os.File) {
	url := listPullURL(repo)
	log.Printf("Crawling %s", url)

	client := http.Client{}

	for {
		resp, err := client.Do(newReq(url, user, passwd))
		if err != nil {
			log.Fatal(err)
		}

		links := parseHeaderLink(resp)

		prs, err := decodePulls(resp)
		if err != nil {
			continue
		}

		for _, pr := range prs {
			fmt.Fprintf(pulls, "%d,%s,%s,%s\n", pr.Number,
				escapeCSV(pr.User.Login), escapeCSV(pr.Title), escapeCSV(pr.Body))

			crawlComments(pr.Number, repo, user, passwd, comments)
		}

		if _, ok := links["next"]; !ok {
			break
		}
		url = links["next"]
		log.Printf("%s\n", url)
	}
}

func crawlComments(number int, repo, user, passwd string, comments *os.File) {
	url := listCommentsURL(repo, number)
	log.Printf("    comment %s", url)

	client := http.Client{}

	for {
		resp, err := client.Do(newReq(url, user, passwd))
		if err != nil {
			log.Fatal(err)
		}

		links := parseHeaderLink(resp)

		cmts, err := decodeComments(resp)
		if err != nil {
			continue
		}

		for _, cmt := range cmts {
			fmt.Fprintf(comments, "%d,%s,%s\n", number,
				escapeCSV(cmt.User.Login), escapeCSV(cmt.Body))
		}

		if _, ok := links["next"]; !ok {
			break
		}
		url = links["next"]
		log.Printf("%s\n", url)
	}
}

func listPullURL(repo string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/pulls?state=all", repo)
}

func listCommentsURL(repo string, number int) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d/comments", repo, number)
}

func newReq(url, user, passwd string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(user) > 0 && len(passwd) > 0 {
		req.SetBasicAuth(user, passwd)
	}
	return req
}

func parseHeaderLink(resp *http.Response) map[string]string {
	links := map[string]string{}

	if _, ok := resp.Header["Link"]; ok {
		headerLink := resp.Header["Link"][0]

		for _, link := range strings.Split(headerLink, ",") {
			ss := strings.Split(link, ";")
			l := betweenAngles.FindStringSubmatch(ss[0])[1]
			t := betweenQuotes.FindStringSubmatch(ss[1])[1]
			links[t] = l
		}
	}
	return links
}

func decodePulls(resp *http.Response) ([]Pull, error) {
	var prs []Pull
	err := json.NewDecoder(resp.Body).Decode(&prs)
	return prs, err
}

func decodeComments(resp *http.Response) ([]Comment, error) {
	var cmts []Comment
	err := json.NewDecoder(resp.Body).Decode(&cmts)
	return cmts, err
}

func escapeCSV(value string) string {
	return strings.Replace(
		strings.Replace(value, ",", " ", -1), // CSV use ,
		"\r\n", " ", -1)                      // GitHub use \r\n for multiline text
}
