package crawler

import "fmt"

const BaseURL = "https://www.linkedin.com/"
const Cookie = "bcookie=\"v=2&bc1f1c97-3916-466a-8bfb-da4be902f68d\"; bscookie=\"v=1&201703121524239cab2092-8525-4763-8364-e41ffb051782AQGJugEZ-xV5NIf1h-r7cf17eU0eLO-a\"; visit=\"v=1&M\"; __ssid=b62dfba2-9c0d-4cdb-a832-9c5b4157e73b; PLAY_SESSION=2519dae9544b537c336b0150f8869acf89317e18-chsInfo=501a88cf-2df3-4139-8a7e-3614c381f37d+premium_nav_upsell_text; cap_session_id=\"2265687193:1\"; _chartbeat2=CxhUlHCReHb-tZw5r.1490519620812.1499461863312.0000000000000001; _ga=GA1.2.535940296.1490482468; lang=\"v=2&lang=fr-fr\"; __utma=226841088.535940296.1490482468.1513295866.1513295866.1; __utmc=226841088; __utmz=226841088.1513295866.1.1.utmcsr=google|utmccn=(organic)|utmcmd=organic|utmctr=(not%20provided); JSESSIONID=\"ajax:3907710712361735591\"; sl=\"v=1&Za0qE\"; li_at=AQEDARLkRUECDABZAAABYSBJMaQAAAFhRFW1pE0APPrGKJIyhT87SVQgYrqf2zRk2wtAhPTgXjGnraUbW8d3wgeStTxpUIWMgxnK7wCRH3U7Q_ASRVA9xIqdZbu8rjAtenqwYuWOt3rs57UUySr3JnMB; liap=true; sdsc=1%3A1SZM1shxDNbLt36wZwCgPgvN58iw%3D; _lipt=CwEAAAFhI6pM2HRfw716p34Uy3OSbcVUemIzh6K_oddLrzodPnOU8JKia85AdQ9WWL965nuCUHF0-E1Ano1beMXRelaNaBCIRrV4QiW6zyxFgRLfED-g8uARDzwyXFJ4IhjH5fE1jMrDUGu5_czpCBIoI2gUOIxjQSVySOtsRbfGOQ_3YG3C-BcG6Yl9-rsAewRKEkvyynQMyetNiOM3Uus-z_PVw0dDJdAM8kzKKEmd4mm4Wss2DVotGDs6JmR2CeqjE_pANeyu2EdV_eyML2e75HcTZE0hxwTLPbN5cIUEnHXIKeqXwbztYS1erTA6fJkwQxcxNoFne1ZX5LK44T5HMaGOoqp_RMAMxsRuXMtyadpn9q3DAbclwUNbGEJ-S99i6Uy15hPvXWU; lidc=\"b=TB25:g=1257:u=147:i=1516740290:t=1516824354:s=AQEAPE50FAiAegLx2di0aDXntlxz8-7k\""
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"

type Crawler struct {
	urlSource     FileTaskGenerator
	workflow      Workflow
	simpleFetcher HttpFetcher
	analyser      LinkedinAnalyser
}

func NewCrawler(urlsFilePath string, linkFilePath string, targetFilePath string, pageSize int, maxPage int)(*Crawler) {
	var crawler Crawler
	crawler.urlSource = FileTaskGenerator{urlsFilePath, pageSize, maxPage}
	crawler.analyser = *NewSimpleAnalyser(linkFilePath, targetFilePath)
	crawler.workflow.tasksChannel = make(chan Task, 3)
	crawler.workflow.resultsChannel = make(chan Task, 3)
	crawler.workflow.dataChannel = make(chan interface{}, 50)
	return &crawler
}

func (crawler *Crawler) Run() {
	tasks := crawler.urlSource.Generate()
	for _, t := range tasks {
		crawler.workflow.tasksChannel <- t
	}

	for {
		select {
		case task := <- crawler.workflow.tasksChannel:
			// create a fetcher to work
			go func() {crawler.workflow.resultsChannel <- crawler.simpleFetcher.Fetch(task)}()
		case result := <- crawler.workflow.resultsChannel:
			// do with the crawler result
			go func() {
				if tasks, targets, err := crawler.analyser.Analyse(result); err == nil {
					fmt.Printf("Successfully analysed the url %s.\n", result.url)
					fmt.Printf("Got new %d tasks ", len(tasks))
					fmt.Printf("Got new %d targets ", len(targets))
					for _, t := range tasks {
						crawler.workflow.tasksChannel <- t
					}
					for _, d := range targets {
						crawler.workflow.dataChannel <- d
					}
				}
			}()
		case data := <- crawler.workflow.dataChannel:
			go func() {
				fmt.Printf("Got data %v", data)
			}()
		default:
			// do nothing

		}
	}
}
