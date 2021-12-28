package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Get Information from file and return the result
func read_file() []byte {
	// "/proc/net/snmp" is the file that contains the information
	content, err := ioutil.ReadFile("/home/snmp")
	if err != nil {
		log.Fatal(err)
	}
	return content
}

// Execute Values from result string
func extract_values() (string, string) {
	res := read_file()
	re := regexp.MustCompile(`\d+`)
	all_numbers := re.FindAllString(string(res), -1)
	segment_sent_out := re.FindAllString(all_numbers[len(all_numbers)-21], -1)[0]
	segment_retrans := re.FindAllString(all_numbers[len(all_numbers)-20], -1)[0]
	return segment_sent_out, segment_retrans
}

// get information and return them
func get_tcp_segment() (int64, int64) {
	segment_sent_out, segment_retrans := extract_values()
	segment_sent_out_integer, _ := strconv.ParseInt(segment_sent_out, 10, 32)
	segment_retrans_integer, _ := strconv.ParseInt(segment_retrans, 10, 32)

	return segment_sent_out_integer, segment_retrans_integer
}

func recordMetrics() {
	go func() {
		for {
			segment_sent_out, segment_retrans := get_tcp_segment()
			tcpSegment.Set(float64(segment_sent_out))
			tcpretransmission.Set(float64(segment_retrans))
			tcpRate.Set(float64(segment_retrans) / float64(segment_sent_out))
			time.Sleep(10 * time.Second)
		}
	}()
}

var (
	tcpSegment = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tcp_segment",
		Help: "The number of TCP segments processed",
	})
	tcpretransmission = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tcp_retransmission",
		Help: "The number of TCP retransmited",
	})
	tcpRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tcp_rate",
		Help: "The number of TCP rate",
	})
)

func main() {
	recordMetrics()
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
