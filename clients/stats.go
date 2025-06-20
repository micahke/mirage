package clients

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

type StatsCounter interface {
	Inc()
}

type StatsClient interface {
	Counter(name string) StatsCounter
	Scope(scopes ...string) StatsClient
}

func StartPromListener(port int) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Fatalf("Failed to start Prometheus listener: %v", err)
		}
	}()
}

var (
	registeredCache = make(map[string]prometheus.Counter)
	cacheMutex      sync.Mutex
)

func fetchCounter(name string) prometheus.Counter {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if counter, ok := registeredCache[name]; ok {
		return counter
	}

	return nil
}

func scopeToName(scopes []string) string {
	return strings.Join(scopes, ":")
}

type StatsV2Client struct {
	scopes []string
}

func NewStatsV2Client(scopes ...string) *StatsV2Client {
	return &StatsV2Client{
		scopes: scopes,
	}
}

func (s *StatsV2Client) Counter(name string) StatsCounter {
	newName := scopeToName(append(s.scopes, name))
	if counter := fetchCounter(newName); counter != nil {
		return counter
	}

	counter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: newName,
			Help: "Some name",
		},
	)

	prometheus.MustRegister(counter)

	cacheMutex.Lock()
	registeredCache[newName] = counter
	cacheMutex.Unlock()

	return counter
}

func (s *StatsV2Client) Scope(scopes ...string) StatsClient {
	return &StatsV2Client{
		scopes: append(s.scopes, scopes...),
	}
}
