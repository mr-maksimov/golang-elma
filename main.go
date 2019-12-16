package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

//K: кол-во буферов в канале
//Template: искомая фраза
const (
	K        = 5
	Template = "Go"
)

type result struct {
	url   string
	count int
	err   error
}

func getCountFindInURL(url string) result {
	var res result
	var resp *http.Response
	var body []byte

	resp, res.err = http.Get(url)
	if res.err != nil {
		return res
	}
	body, res.err = ioutil.ReadAll(resp.Body)
	if res.err != nil {
		return res
	}

	if res.err == nil {
		res = result{
			url:   url,
			count: strings.Count(string(body), Template),
		}
	}
	return res
}

func sendMsg(urls <-chan result, done chan<- string) {
	var sum int
	for url := range urls {
		if url.err != nil {
			fmt.Println(url.err)
			return
		}
		sum += url.count
		fmt.Printf("Count '%v' for %s: %d\n", Template, url.url, url.count)
	}
	fmt.Printf("Total: %d\n", sum)
	done <- "I'm done"
}

func main() {
	_, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	urls := make(chan result)
	done := make(chan string)

	go sendMsg(urls, done)

	scanner := bufio.NewScanner(os.Stdin)
	var wg sync.WaitGroup
	buffer := make(chan string, K)
	for scanner.Scan() {
		url := scanner.Text()
		buffer <- "buffer"
		wg.Add(1)
		go func(url string, buffer <-chan string, urls chan<- result, wg *sync.WaitGroup) {
			urls <- getCountFindInURL(url)
			<-buffer
			wg.Done()
		}(url, buffer, urls, &wg)
	}
	wg.Wait()
	close(urls)
	close(buffer)
	<-done
}
