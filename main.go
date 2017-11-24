package main

import (
  "log"
	"flag"
	"os"
	"io/ioutil"
//	"log"
	"net/http"
  "fmt"
	"mvdan.cc/xurls"
	_ "reflect"
  "strings"
	"time"
  "sync"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	fpc_load_time = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fpc_load_time",
		Help: "Time to load the landing page in seconds.",
	})
	fpc_load_failures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fpc_load_failures",
			Help: "Number of times we failed to load a page.",
		},
		[]string{"device"},
	)
)

func init() {
	prometheus.MustRegister(cpuTemp)
	prometheus.MustRegister(hdFailures)
}

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func getContents(url string) (string, err) {
    var contents []byte
    fmt.Printf("getting contents for: %s\n", url)
    response, err := http.Get(url)
		if err != nil {
				fmt.Printf("%s", err)
		} else {
				defer response.Body.Close()
				contents, err = ioutil.ReadAll(response.Body)
				if err != nil {
						fmt.Printf("%s", err)
				}
				//fmt.Printf("%s\n", string(contents))
		}
		return string(contents), err
}

func checkUrl(url string) {
  //fmt.Printf("-> doing %s\n", url)
}

func doStuff(url string) {
  time_start = time.Now()
  c, err := getContents(url)
  if err != nil {

  }
  else {

  }
  childUrls := xurls.Strict().FindAllString(c, -1)
  //fmt.Printf("%v", childUrls)
  for _, childUrl := range childUrls {
    if strings.Contains(childUrl, "http://www.w3.org") {
      fmt.Printf("X ignoring %s\n", childUrl)
    } else {
      checkUrl(childUrl)
    }
  }
}

func startChecking(urls []string) {
  for {
    const workers = 25

    wg := new(sync.WaitGroup)
    in := make(chan string, 2*workers)

    for i:= 0; i < workers; i++ {
      wg.Add(1)
      go func() {
        defer wg.Done()
        for url := range in {
          doStuff(url)
        }
      }()
    }

    for _, url := range urls {
      in <- url
    }
    close(in)
    wg.Wait()

    fmt.Println("Sleeping...")
    time.Sleep(5 * time.Second)
  }
}

func main() {
	flag.Parse()

  urls := []string{}
  urls = append(urls, "http://www.wp.pl")

  go startChecking(urls)

  cpuTemp.Set(65.3)
	hdFailures.With(prometheus.Labels{"device":"/dev/sda"}).Inc()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))

}
