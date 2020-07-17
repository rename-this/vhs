package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 1024,
		TLSHandshakeTimeout: 0 * time.Second,
	}
	client := &http.Client{Transport: tr}

	for {
		data := strings.NewReader(fmt.Sprintf("hello, world %d", time.Now().UnixNano()))
		req, err := http.NewRequest("POST", "http://localhost:1111", data)
		if err != nil {
			panic(err)
		}
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		io.Copy(os.Stdout, res.Body)
		fmt.Println()
		time.Sleep(time.Second)
	}
}
