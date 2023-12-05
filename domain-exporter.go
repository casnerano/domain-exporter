package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type domain struct {
	name     string
	paidTill time.Time
	freeDate time.Time
}

var addr = flag.String("a", "localhost:8080", "Server address")

func main() {
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := getIndexTemplate().Execute(w, nil); err != nil {
			log.Println("Failed render index page.")
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target := strings.TrimSpace(r.URL.Query().Get("target"))
		if target == "" {
			w.WriteHeader(http.StatusBadRequest)
			// ......
		}
	})

	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func whois(host string) *domain {
	out, err := exec.Command("bash", "-c", fmt.Sprintf("whois %s", host)).Output()
	if err != nil {
		log.Fatal(err)
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
			if pt, ptErr := time.Parse(time.RFC3339, value); err == nil {
				d.paidTill = pt
			} else {
				log.Printf("Failed parse paid-till date for %s: %s\n", host, ptErr.Error())
			}
		case "free-date:":
			if fd, fdErr := time.Parse(time.RFC3339, value); err == nil {
				d.paidTill = fd
			} else {
				log.Printf("Failed parse paid-till date for %s: %s\n", host, fdErr.Error())
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
					<p><a href="/metrics"></a></p>
					<p><a href="/probe?target=ya.ru">probe ya.ru</a></p>
				</body>
			</html>`,
		),
	)
}
