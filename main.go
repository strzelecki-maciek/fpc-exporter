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
	_ "time"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_temperature_celsius",
		Help: "Current temperature of the CPU.",
	})
	hdFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hd_errors_total",
			Help: "Number of hard-disk errors.",
		},
		[]string{"device"},
	)
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func getContents(url string) string {
    var contents []byte
    fmt.Printf("getting contents for: %s\n", url)
    response, err := http.Get(url)
		if err != nil {
				fmt.Printf("%s", err)
				os.Exit(1)
		} else {
				defer response.Body.Close()
				contents, err = ioutil.ReadAll(response.Body)
				if err != nil {
						fmt.Printf("%s", err)
						os.Exit(1)
				}
				//fmt.Printf("%s\n", string(contents))
		}
		return string(contents)
}

func checkUrl(url string) {
  fmt.Printf("-> doing %s\n", url)
}

func init() {
	prometheus.MustRegister(cpuTemp)
	prometheus.MustRegister(hdFailures)
}

func main() {
	flag.Parse()

  urls := []string{}
  urls = append(urls, "http://www.wp.pl")

  for _, url := range urls {
	  childUrls := xurls.Strict().FindAllString(getContents(url), -1)
		fmt.Printf("%v", childUrls)
		for _, childUrl := range childUrls {
			if strings.Contains(childUrl, "http://www.w3.org") {
				fmt.Printf("X ignoring %s\n", childUrl)
			} else {
			  checkUrl(childUrl)
			}
    }
  }

  cpuTemp.Set(65.3)
	hdFailures.With(prometheus.Labels{"device":"/dev/sda"}).Inc()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))

}
