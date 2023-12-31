package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "domain"
)

type (
	domain = string
	result struct {
		paidTill time.Time
		freeDate time.Time
	}
)

var (
	descriptions = map[string]*prometheus.Desc{
		"success": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "success"),
			"Domain check was successful",
			[]string{"domain"},
			nil,
		),
		"paid_till": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "paid_till_seconds"),
			"Domain paid till",
			[]string{"domain"},
			nil,
		),
		"free_date": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "free_date_seconds"),
			"Domain free date",
			[]string{"domain"},
			nil,
		),
	}
)

type collector struct {
	domain domain
}

func newCollector(domain domain) *collector {
	return &collector{
		domain: domain,
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range descriptions {
		ch <- desc
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var success, paidTill, freeDate float64

	res, err := whois(ctx, c.domain)
	if err == nil {
		success = 1
		paidTill = res.paidTill.Sub(time.Now()).Seconds()
		freeDate = res.freeDate.Sub(time.Now()).Seconds()
	}

	log.Printf(
		"Collected for %s:  (success: %v; paid_till_seconds: %v; free_date_seconds: %v)\n",
		c.domain,
		success,
		paidTill,
		freeDate,
	)

	ch <- prometheus.MustNewConstMetric(descriptions["success"], prometheus.GaugeValue, success, c.domain)
	ch <- prometheus.MustNewConstMetric(descriptions["paid_till"], prometheus.GaugeValue, paidTill, c.domain)
	ch <- prometheus.MustNewConstMetric(descriptions["free_date"], prometheus.GaugeValue, freeDate, c.domain)
}

var addr = flag.String("a", ":80", "server address")
var debug = flag.Bool("d", false, "debug mode")

func main() {
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := getIndexTemplate().Execute(w, nil); err != nil {
			log.Println("Failed render index page.")
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		target := strings.TrimPrefix(
			strings.TrimSpace(r.URL.Query().Get("target")), "www.",
		)

		if target == "" {
			w.WriteHeader(http.StatusBadRequest)
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(newCollector(target))

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})

	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func whois(ctx context.Context, d domain) (*result, error) {
	out, err := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("whois %s", d)).Output()
	if err != nil {
		return nil, err
	}

	res := result{}

	if *debug {
		log.Printf("whois %s\n%s\n", d, out)
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}

		key := strings.TrimRight(strings.TrimSpace(fields[0]), ":")
		value := strings.TrimSpace(fields[1])

		switch key {
		case "paid-till":
			res.paidTill, _ = time.Parse(time.RFC3339, value)
		case "free-date":
			res.freeDate, _ = time.Parse(time.DateOnly, value)
		}
	}

	if res.paidTill.IsZero() || res.freeDate.IsZero() {
		return nil, fmt.Errorf("failed values parse")
	}

	return &res, nil
}

func getIndexTemplate() *template.Template {
	return template.Must(
		template.New("index").Parse(
			`<html>
				<head>
					<title>Domain Exporter</title>
				</head>
				<body>
					<h1>Domain Exporter</h1>
					<p><a href="/metrics">Metrics</a></p>
					<p><a href="/probe?target=ya.ru">Probe ya.ru</a></p>
				</body>
			</html>`,
		),
	)
}
