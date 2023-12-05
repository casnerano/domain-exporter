package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
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
		w.Header().Set("Content-Type", "text/plain")
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
