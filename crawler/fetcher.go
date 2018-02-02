package crawler

import (
	"bufio"
	"log"
	"os"
	"net/http"
	"io/ioutil"
	"fmt"
)

type Task struct {
	url string
	depth int
	body string
	statusCode int
	err error
}

func NewTask(url string, depth int) *Task {
	return &Task{url:url, depth:depth, body:"", statusCode:0, err: nil}
}

type Workflow struct {
	tasksChannel   chan Task
	resultsChannel chan Task
	dataChannel chan interface{}	// a channel containing the instance to be persisted in database
}

type HttpFetcher struct {

}

func (fetcher *HttpFetcher) Fetch(action Task) (Task) {
	client := http.Client{}
	req, err := http.NewRequest("GET", action.url, nil)
	req.Header.Set("cookie", Cookie)
	req.Header.Set("user-agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return Task{url: action.url, err: err}
	}
	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		return Task{url: action.url, depth: action.depth - 1, body: string(b), statusCode: resp.StatusCode}
	} else {
		return Task{url: action.url, err: err}
	}
}

type FileTaskGenerator struct {
	urlsFilePath string
	pageSize int
	maxPage int
}

func (g FileTaskGenerator) Generate() ([]Task) {
	file, err := os.Open(g.urlsFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var actions []Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		for page := 0; page < g.pageSize * g.maxPage; page += g.pageSize {
			actions = append(actions, *NewTask(scanner.Text() + "&start=" + string(page), 1))
		}
	}
	fmt.Printf("A list of %d task is generated.\n", len(actions))
	return actions
}
