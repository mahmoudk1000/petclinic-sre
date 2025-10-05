package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultAppUrl       = "http://petclinic-prd.prd:8080"
	defaultExporterPort = "9098"
	appMetricsEndpoint  = "/actuator/prometheus"
)

var (
	appUrl       = getEnv("APP_URL", defaultAppUrl)
	appEndpoint  = getEnv("APP_ENDPOINT", appMetricsEndpoint)
	exporterPort = getEnv("EXPORTER_PORT", defaultExporterPort)
	metricRegex  = regexp.MustCompile(
		`^(?P<name>[a-zA-Z_:][a-zA-Z0-9_:]*)\{(?P<labels>[^\}]*)\} (?P<value>[0-9\.eE+-]+)$`,
	)
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func fetchMetrics() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s", appUrl, appEndpoint))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func transformMetrics(metrics string) string {
	var result strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(metrics))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# HELP ") || strings.HasPrefix(line, "# TYPE ") {
			// Modify the HELP and TYPE lines
			lineParts := strings.Split(line, " ")
			lineParts[2] = "petclinic_" + lineParts[2]
			result.WriteString(strings.Join(lineParts, " ") + "\n")
		} else if len(line) > 0 {
			// Prepend "petclinic_" to the metric name
			spaceIndex := strings.Index(line, " ")
			if spaceIndex > 0 {
				result.WriteString("petclinic_" + line[:spaceIndex] + line[spaceIndex:] + "\n")
			} else {
				result.WriteString(line + "\n")
			}
		}
	}

	return result.String()
}

type customCollector struct {
	desc *prometheus.Desc
}

func newCustomCollector() *customCollector {
	return &customCollector{
		desc: prometheus.NewDesc(
			"petclinic_metrics",
			"Custom metrics from the PetClinic application",
			nil,
			nil,
		),
	}
}

func (c *customCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c *customCollector) Collect(ch chan<- prometheus.Metric) {
	metrics, err := fetchMetrics()
	if err != nil {
		log.Printf("Error fetching metrics: %v", err)
		return
	}

	transformedMetrics := transformMetrics(metrics)
	scanner := bufio.NewScanner(strings.NewReader(transformedMetrics))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			match := metricRegex.FindStringSubmatch(line)
			if match != nil {
				metricName := match[1]
				labelsStr := match[2]
				valueStr := match[3]

				labels := parseLabels(labelsStr)
				value, err := strconv.ParseFloat(valueStr, 64)
				if err == nil {
					ch <- prometheus.MustNewConstMetric(
						prometheus.NewDesc(metricName, "", labels.names, nil),
						prometheus.GaugeValue, value, labels.values...,
					)
				}
			}
		}
	}
}

type labelSet struct {
	names  []string
	values []string
}

func parseLabels(labelsStr string) labelSet {
	labels := labelSet{}
	pairs := strings.Split(labelsStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			labels.names = append(labels.names, kv[0])
			labels.values = append(labels.values, strings.Trim(kv[1], `"`))
		}
	}
	return labels
}

func main() {
	// Create a new registry.
	reg := prometheus.NewRegistry()

	// Register the custom collector.
	customCol := newCustomCollector()
	reg.MustRegister(customCol)

	// Expose the registered metrics at `/metrics` endpoint.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Printf("Starting exporter at port %s", exporterPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", exporterPort), nil))
}
