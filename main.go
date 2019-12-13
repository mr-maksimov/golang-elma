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

func getCountFindInURL(url string) (count int, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return strings.Count(string(body), Template), nil
}

func sendMsg(urls chan string, done chan string) {
	var sum int
	for url := range urls {
		count, err := getCountFindInURL(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		sum += count
		fmt.Printf("Count '%v' for %s: %d\n", Template, url, count)
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

	urls := make(chan string)
	done := make(chan string)

	go sendMsg(urls, done)

	scanner := bufio.NewScanner(os.Stdin)
	var wg sync.WaitGroup
	buffer := make(chan string, K)
	for scanner.Scan() {
		url := scanner.Text()
		buffer <- "buffer"
		wg.Add(1)
		go func(url string, buffer <-chan string, urls chan<- string, wg *sync.WaitGroup) {
			urls <- url
			<-buffer
			wg.Done()
		}(url, buffer, urls, &wg)
	}
	wg.Wait()
	close(urls)
	close(buffer)
	<-done
}
