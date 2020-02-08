// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed
// with this work for additional information regarding copyright
// ownership.  The ASF licenses this file to you under the Apache
// License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License.  You may obtain a copy of
// the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.  See the License for the specific language governing
// permissions and limitations under the License.

// To build this crawler, you need to install Go and run the following
// command.
//
//   go install ./...
//
// To crawl the pull requests and comments of all pull requests,
// please run the following command.
//
//  $GOPATH/bin/crawler -user your_github_account -passwd your_github_passwd
//
// Or, you can set your username and password in environment variables.
//
//  export GITHUB_USER=your_github_account
//  export GITHUB_PASSWD=your_github_passwd
//  $GOPATH/bin/crawler
//
// This crawler generates two CSV files: pulls.csv and comments.csv.
// Both have a column named number, which is the pull request number.
//
// You can import these two files into a SQL database, for example,
// MySQL, as two tables, you can join these two tables by number.  For
// more about the importing and NLP with MySQL, please refer to
// https://medium.com/@yi.wang.2005/nlp-in-sql-word-vectors-82dffc908423
//
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
