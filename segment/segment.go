package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xujiajun/gotokenizer"
)

func main() {
	mm := gotokenizer.NewMaxMatch("dict.txt")
	if e := mm.LoadDict(); e != nil {
		log.Fatal(e)
	}

	mm.EnabledFilterStopToken = true
	mm.StopTokens = gotokenizer.NewStopTokens()
	mm.StopTokens.Load("stop_tokens.txt")

	f, e := os.Open(os.Args[1])
	if e != nil {
		log.Fatal(e)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		t = strings.ReplaceAll(t, "\t", " ")
		t = strings.ReplaceAll(t, "\r", " ")
		if len(t) > 10 {
			s, e := mm.GetFrequency(t)
			if e != nil {
				log.Fatal(e)
			}
			if len(s) > 0 {
				for k, v := range s {
					fmt.Printf("%s %d ", k, v)
				}
				fmt.Println()
			}
		}
	}
	if e := scanner.Err(); e != nil {
		log.Fatal(e)
	}
}
