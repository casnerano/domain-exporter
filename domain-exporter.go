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

var addr = flag.String("a", ":80", "server address")

type domain struct {
	name     string
	paidTill time.Time
}

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
		target := strings.TrimSpace(r.URL.Query().Get("target"))
		if target == "" {
			w.WriteHeader(http.StatusBadRequest)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		d := whois(ctx, target)

		dPaidTill := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "paid_till_seconds",
				Help: "Domain paid till seconds",
			},
			[]string{"domain"},
		)

		registry := prometheus.NewRegistry()
		registry.MustRegister(dPaidTill)

		dPaidTill.With(prometheus.Labels{"domain": d.name}).Set(d.paidTill.Sub(time.Now()).Seconds())

		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})

	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func whois(ctx context.Context, host string) *domain {
	out, err := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("whois %s", host)).Output()
	if err != nil {
		log.Println(err)
	}

	d := domain{name: host}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}

		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "paid-till:":
			if pt, ptErr := time.Parse(time.RFC3339, value); ptErr == nil {
				d.paidTill = pt
			} else {
				log.Printf("Failed parse paid-till date for %s: %s\n", host, ptErr.Error())
			}
		}
	}

	return &d
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
					<p><a href="/metrics">metrics</a></p>
					<p><a href="/probe?target=ya.ru">probe ya.ru</a></p>
				</body>
			</html>`,
		),
	)
}
