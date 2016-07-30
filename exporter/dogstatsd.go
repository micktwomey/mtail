package exporter

import (
	"expvar"
	"flag"
	"fmt"
	"strings"

	"github.com/google/mtail/metrics"
)

var (
	dogstatsdHostPort = flag.String("dogstatsd_hostport", "",
		"Host:port to dogstatsd server to write metrics to.")

	dogstatsdExportTotal   = expvar.NewInt("dogstatsd_export_total")
	dogstatsdExportSuccess = expvar.NewInt("dogstatsd_export_success")
)

func stripInvalidCharacters(s string) string {
	var replacements = []string{":", ",", " ", "%", "(", ")"}
	for _, original := range replacements {
		s = strings.Replace(s, original, "_", -1)
	}
	return s
}

func formatLabelsToTags(m map[string]string, ksep, sep string) string {
	if len(m) > 0 {
		var s []string
		for k, v := range m {
			s = append(s, fmt.Sprintf("%s%s%s", k, ksep, v))
		}
		return strings.Join(s, sep)
	}
	return ""
}

func metricToDogstatsd(hostname string, m *metrics.Metric, l *metrics.LabelSet) string {
	m.RLock()
	defer m.RUnlock()
	var t string
	switch m.Kind {
	case metrics.Counter:
		t = "c" // StatsD Counter
	case metrics.Gauge:
		t = "g" // StatsD Gauge
	case metrics.Timer:
		t = "ms" // StatsD Timer
	}
	var labels = make(map[string]string)
	labels["program"] = stripInvalidCharacters(m.Program)
	for key, value := range l.Labels {
		labels[stripInvalidCharacters(key)] = stripInvalidCharacters(value)
	}
	return fmt.Sprintf("%s:%d|%s|#%s",
		m.Name,
		l.Datum.Get(),
		t,
		formatLabelsToTags(labels, ":", ","))
}
