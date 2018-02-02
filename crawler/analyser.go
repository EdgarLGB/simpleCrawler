package crawler

import (
	"fmt"
	"errors"
	"github.com/ChrisTrenkamp/goxpath"
	"github.com/ChrisTrenkamp/goxpath/tree/xmltree"
	"bytes"
	"github.com/ChrisTrenkamp/goxpath/tree"
	"os"
	"log"
	"bufio"
	"strings"
	"strconv"
	"net/http"
)

type Analyser interface {
	Analyse(result Task) (task[]Task, taskCompleted[]Task, err error)
}

type LinkedinAnalyser struct {
	linkXPaths   map[string]int //The link in the page
	targetXPaths []string       //Xpath of the target object
}

func NewSimpleAnalyser(linkFilePath string, targetFilePath string) *LinkedinAnalyser {
	var analyser LinkedinAnalyser
	// Read the file which defines the extraction of the url
	if file, err := os.Open(linkFilePath); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			entry := strings.Split(scanner.Text(), ":")
			analyser.linkXPaths[strings.Trim(entry[1], " ")], err = strconv.Atoi(strings.Trim(entry[0], " "))
			if err != nil {
				log.Fatal(fmt.Sprintf("Cannot parse the url extraction text file. %v", err))
			}
		}
		file.Close()
	} else {
		log.Fatal(fmt.Sprintf("File %s can't be opened. Because of %v", file.Name(), err))
	}
	// Read the file which defines the xPath of target object
	if file, err := os.Open(targetFilePath); err == nil{
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			analyser.targetXPaths = append(analyser.targetXPaths, scanner.Text())
		}
		file.Close()
	} else {
		log.Fatal(fmt.Sprintf("File %s can't be opened. Because of %v", file.Name(), err))
	}
	return &analyser
}

func (analyser *LinkedinAnalyser)Analyse(result Task) ([]Task, []interface{}, error) {
	if result.statusCode != http.StatusOK {
		return nil, nil, errors.New(fmt.Sprintf("%s got the status code %d", result.url, result.statusCode))
	}
	if result.err != nil {
		return nil, nil, errors.New(fmt.Sprintf("%s got the error %v", result.url, result.err))
	}
	//Needs to add an extractor to get the useful information
	targets, err1 := analyser.extractTarget(result)
	//Needs to extract the url
	tasks, err2 := analyser.extractURL(result.body)
	if err1 != nil {
		return tasks, nil, errors.New(fmt.Sprintf("%s got the error when extracting the targets %v", result.url, err1))
	}
	if err2 != nil {
		return nil, targets, errors.New(fmt.Sprintf("%s got the error when extracting the urls %v", result.url, err2))
	}
	return tasks, targets, nil
}

func (analyser *LinkedinAnalyser) extractTarget(data Task) ([]interface{}, error) {
	var result []interface{}
	xTree, err := xmltree.ParseXML(bytes.NewBufferString(data.body))
	if err != nil{
		return nil, err
	}
	for _, v := range analyser.targetXPaths {
		var xpExec = goxpath.MustParse(v)
		res, err := xpExec.ExecNode(xTree)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(" cannot extract the target from this page %s because of %v", data.url, err))
		}
		viewSize, _ := strconv.Atoi(strings.Split(res[4].ResValue(), " ")[0])
		j := job{res[0].ResValue(), res[1].ResValue(), res[2].ResValue(), res[3].ResValue(), viewSize, res[5].ResValue(), res[6].ResValue()}
		result = append(result, j)
	}
	return result, nil
}

func (analyser *LinkedinAnalyser) extractURL(body string) ([]Task, error) {
	xTree, err := xmltree.ParseXML(bytes.NewBufferString(body))
	if err != nil{
		return nil, err
	}
	var result []Task
	for k, v := range analyser.linkXPaths {
		var xpExec = goxpath.MustParse(k)
		res, err := xpExec.ExecNode(xTree)
		if err != nil {
			// skip this xPath
			continue
		}
		for _, n := range res {
			if elem, ok := n.(tree.Elem); ok {
				for _, attr := range elem.GetAttrs() {
					// Need to get the attribute "href"
					if attr.GetToken() == "href" {
						result = append(result, *NewTask(BaseURL+ attr.ResValue(), v))
					}
				}
			}
		}
	}
	return result, nil
}

