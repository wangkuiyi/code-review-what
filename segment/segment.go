package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xujiajun/gotokenizer"
)

func main() {
	stopwords := flag.String("stopwords", path.Join(filepath.Dir(os.Args[0]), "stop_tokens.txt"), "stopword file")
	dict := flag.String("dict", path.Join(filepath.Dir(os.Args[0]), "dict.txt"), "dictionary file")
	flag.Parse()

	mm := gotokenizer.NewMaxMatch(*dict)
	if e := mm.LoadDict(); e != nil {
		log.Fatal(e)
	}

	mm.EnabledFilterStopToken = true
	mm.StopTokens = gotokenizer.NewStopTokens()
	mm.StopTokens.Load(*stopwords)

	f, e := os.Open(os.Args[1])
	if e != nil {
		log.Fatal(e)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		t = strings.ReplaceAll(t, "\t", " ")
		t = strings.ReplaceAll(t, "\r", " ")

		s, e := mm.Get(t)
		if e != nil {
			log.Fatal(e)
		}

		if len(s) >= 2 {
			fmt.Printf("%s\n", strings.Join(s, "\t"))
		}
	}
	if e := scanner.Err(); e != nil {
		log.Fatal(e)
	}
}
