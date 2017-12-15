package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	hash_check = "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="
)

func TestMain(m *testing.M) {
	go func() {
		server()
	}()
	// Allow server to start
	delay()
	os.Exit(m.Run())
}

func TestHash(t *testing.T) {
	println()
	log.Println("Testing single hash...")
	res, dur := postPassword("angryMonkey", false)

	if res != hash_check {
		log.Println(res, hash_check)
		t.Error("Incorrect hash recieved")
	} else if dur < 5*time.Second {
		t.Error("Response too fast")
	}
	log.Println("Correct Hash recieved in:", dur)
}

func TestMultiRequest(t *testing.T) {
	connections := 100
	println()
	log.Println("Testing", connections, "requests in parallel ...")

	var wg sync.WaitGroup
	wg.Add(connections)
	for i := 0; i < connections; i++ {
		log.Println("Starting request", i)
		go func() {
			defer wg.Done()
			res, dur := postPassword("angryMonkey", false)
			if res != hash_check {
				t.Error("Incorrect hash recieved")
			} else if dur < 5*time.Second {
				t.Error("Response too fast")
			}
		}()
	}
	wg.Wait()
	log.Println("All requests answered quickly")
}

func TestShutdown(t *testing.T) {
	println()
	log.Println("Testing graceful shutdown ...")
	done := make(chan bool)
	go func() {
		log.Println("Request1 send")
		res, dur := postPassword("angryMonkey", false)
		if res != hash_check {
			t.Error("Incorrect hash recieved")
		} else if dur < 5*time.Second {
			t.Error("Response too fast")
		}
		log.Println("Recieved hash from request1")
		close(done)
	}()
	// Give goroutine a chance to start
	delay()

	resp, err := http.Get("http://localhost:8080/shutdown")
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(body) != "Performing graceful shutdown" {
		t.Error("Did not recieve correct shutdown response")
	}
	log.Println("Request2 send")
	_, _ = postPassword("angryMonkey", true)
	<-done
}

func postPassword(password string, supposedToFail bool) (string, time.Duration) {
	start := time.Now()
	resp, err := http.PostForm("http://localhost:8080/", url.Values{"password": {password}})
	if err != nil {
		if supposedToFail {
			log.Println("Post failed, this was the correct behavior")
			return "", time.Hour
		}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	dur := time.Since(start)
	return string(body), dur
}

func delay() {
	time.Sleep(500 * time.Millisecond)
}
