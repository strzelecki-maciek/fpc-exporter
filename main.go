package main

import (
	"flag"
  _ "os"
	"io/ioutil"
  "errors"
	"net/http"
  "fmt"
	"mvdan.cc/xurls"
	_ "reflect"
  "strconv"
  "strings"
	"time"
  "sync"
  "os"
  "log"
  "encoding/json"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
)


type Target struct {
  Agent string `json:"agent"`
  URL string `json:"url"`
  Parent string
}

type Configuration struct {
  Targets []Target `json:"targets"`
  QueryInterval int64 `json"queryInterval"`
}

type ContentsResult struct {
  err error
  statusCode string
  contents string
}

var (
	fpcLoadTime = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Name: "fpc_load_time",
      Help: "Time to load the landing page in seconds.",
    },
    []string{"page", "statusCode", "parent"},
  )
  fpcLoadFailure = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fpc_load_failures",
			Help: "Number of times we failed to load a page.",
		},
		[]string{"page", "statusCode", "parent"},
	)
)

func init() {
	prometheus.MustRegister(fpcLoadTime)
	prometheus.MustRegister(fpcLoadFailure)
}

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
var configPath = flag.String("config-path", "config.json", "The path to config file (default: config.json).")

func loadConfig(path string) (Configuration, error) {
  file, _ := os.Open(path)
  defer file.Close()

  decoder := json.NewDecoder(file)
  configuration := Configuration{}
  err := decoder.Decode(&configuration)
  return configuration, err
}

func getContents(url Target) (string, string, error) {
    var contents []byte
    statusCode := "0"
    client := &http.Client{}

    c1 := make(chan ContentsResult, 1)
    go func() {
      //fmt.Println("Trying: " + url.URL + "\n")
      req, err := http.NewRequest("GET", url.URL, nil)
      if err != nil {
          fmt.Printf("%s", err)
      } else {
        req.Header.Set("User-Agent", url.Agent)
        response, err := client.Do(req)
        if err != nil {
            fmt.Printf("%s", err)
        } else {
            defer response.Body.Close()
            statusCode = strconv.Itoa(response.StatusCode)
            contents, err = ioutil.ReadAll(response.Body)
            if err != nil {
                fmt.Printf("%s", err)
            }
            //fmt.Printf("%s\n", string(contents))
        }
      }
      c1 <- ContentsResult{err:err, statusCode:statusCode,contents:string(contents)}
    }()

    select {
      case res:= <-c1:
        return res.contents, res.statusCode, res.err
      case <-time.After(time.Second * 15):
        return "", "0", errors.New("timeouted")
    }
}

func checkPage(target Target) (string) {
  time_start := time.Now()
  if target.Parent == "" {
    target.Parent = "none"
  }
  c, statusCode, err := getContents(target)
  if statusCode != "200" || err != nil {
    fpcLoadFailure.With(prometheus.Labels{"page": target.URL, "statusCode": statusCode, "parent": target.Parent}).Inc()
  }
  fpcLoadTime.With(prometheus.Labels{"page": target.URL, "statusCode": statusCode, "parent": target.Parent}).Set(time.Since(time_start).Seconds())
  return c
}


func doStuff(parent Target) {
  parentContents := checkPage(parent)
  childUrls := xurls.Strict().FindAllString(parentContents, -1)
  for _, childUrl := range childUrls[:10] {
    if !strings.Contains(childUrl, "http://www.w3.org") && childUrl != parent.URL && childUrl != "" {
    checkPage(Target{URL:childUrl, Agent:parent.Agent, Parent:parent.URL})
    }
  }
}

func startChecking(configuration Configuration) {
  for {
    const workers = 25

    wg := new(sync.WaitGroup)
    in := make(chan Target, 2*workers)

    for i:= 0; i < workers; i++ {
      wg.Add(1)
      go func() {
        defer wg.Done()
        for url := range in {
          doStuff(url)
        }
      }()
    }

    for _, url := range configuration.Targets {
      in <- url
    }
    close(in)
    wg.Wait()

    time.Sleep(time.Duration(configuration.QueryInterval) * time.Second)
  }
}

func main() {
	flag.Parse()

  configuration, err := loadConfig(*configPath)
  if err != nil {
    fmt.Println("Error :", err)
    os.Exit(1)
  }

  go startChecking(configuration)
	http.Handle("/metrics", promhttp.Handler())
  fmt.Println("Starting to listen on: ", *addr)
  fmt.Println("Query interval: ", configuration.QueryInterval, "seconds.")
	log.Fatal(http.ListenAndServe(*addr, nil))

}
