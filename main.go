// Fetchall fetches URLs in parallel and reports their times and sizes.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var out = flag.String("out", "results.log", "file to output program results")

func main() {
	flag.Parse()
	fname := *out
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	start := time.Now()
	fmt.Fprintf(f, "START: %v\n", start)
	ch := make(chan string)
	for _, url := range os.Args[1:] {
		go fetch(url, ch) // start a goroutine
	}
	for range os.Args[1:] {
		fmt.Fprintln(f, <-ch) // receive from channel ch
	}
	fmt.Fprintf(f, "%.2fs elapsed\n", time.Since(start).Seconds())
}

func fetch(url string, ch chan<- string) {
	start := time.Now()
	timeout := time.After(20 * time.Second)
	errc := make(chan string)
	successc := make(chan string)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			errc <- fmt.Sprint(err) // send to channel ch
			return
		}
		nbytes, err := io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close() // don't leak resources
		if err != nil {
			errc <- fmt.Sprintf("while reading %s: %v", url, err)
			return
		}
		secs := time.Since(start).Seconds()
		successc <- fmt.Sprintf("%.2fs %7d %s", secs, nbytes, url)
	}()
	select {
	case errstr := <-errc:
		ch <- errstr
	case successstr := <-successc:
		ch <- successstr
	case <-timeout:
		ch <- "timed out after 20s"
	}
}
