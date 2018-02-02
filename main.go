package main

import "simpleCrawler/crawler"

func main() {
	urlsFilePath := "C:\\Users\\bobo\\go\\src\\simpleCrawler\\url_list.txt"
	linkFilePath := "C:\\Users\\bobo\\go\\src\\simpleCrawler\\url_extract.txt"
	targetFilePath := "C:\\Users\\bobo\\go\\src\\simpleCrawler\\target_xpath.txt"
	myCrawler := crawler.NewCrawler(urlsFilePath, linkFilePath, targetFilePath, 20, 200)
	myCrawler.Run()
}