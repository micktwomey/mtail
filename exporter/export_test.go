// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package exporter

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/mtail/metrics"
	"github.com/kylelemons/godebug/pretty"
)

func FakeSocketWrite(f formatter, m *metrics.Metric) []string {
	var ret []string
	lc := make(chan *metrics.LabelSet)
	go m.EmitLabelSets(lc)
	for l := range lc {
		ret = append(ret, f("gunstar", m, l))
	}
	sort.Strings(ret)
	return ret
}

func TestMetricToCollectd(t *testing.T) {
	ts, terr := time.Parse("2006/01/02 15:04:05", "2012/07/24 10:14:00")
	if terr != nil {
		t.Errorf("time parse error: %s", terr)
	}
	ms := metrics.NewStore()

	scalarMetric := metrics.NewMetric("foo", "prog", metrics.Counter)
	d, _ := scalarMetric.GetDatum()
	d.Set(37, ts)
	ms.Add(scalarMetric)

	r := FakeSocketWrite(metricToCollectd, scalarMetric)
	expected := []string{"PUTVAL \"gunstar/mtail-prog/counter-foo\" interval=60 1343124840:37\n"}
	diff := pretty.Compare(r, expected)
	if len(diff) > 0 {
		t.Errorf("String didn't match:\n%s", diff)
	}

	dimensionedMetric := metrics.NewMetric("bar", "prog", metrics.Gauge, "label")
	d, _ = dimensionedMetric.GetDatum("quux")
	d.Set(37, ts)
	d, _ = dimensionedMetric.GetDatum("snuh")
	d.Set(37, ts)
	ms.ClearMetrics()
	ms.Add(dimensionedMetric)

	r = FakeSocketWrite(metricToCollectd, dimensionedMetric)
	expected = []string{
		"PUTVAL \"gunstar/mtail-prog/gauge-bar-label-quux\" interval=60 1343124840:37\n",
		"PUTVAL \"gunstar/mtail-prog/gauge-bar-label-snuh\" interval=60 1343124840:37\n"}
	diff = pretty.Compare(r, expected)
	if len(diff) > 0 {
		t.Errorf("String didn't match:\n%s", diff)
	}

	timingMetric := metrics.NewMetric("foo", "prog", metrics.Timer)
	d, _ = timingMetric.GetDatum()
	d.Set(123, ts)
	ms.Add(timingMetric)

	r = FakeSocketWrite(metricToCollectd, timingMetric)
	expected = []string{"PUTVAL \"gunstar/mtail-prog/gauge-foo\" interval=60 1343124840:123\n"}
	diff = pretty.Compare(r, expected)
	if len(diff) > 0 {
		t.Errorf("String didn't match:\n%s", diff)
	}
}

func TestMetricToGraphite(t *testing.T) {
	ts, terr := time.Parse("2006/01/02 15:04:05", "2012/07/24 10:14:00")
	if terr != nil {
		t.Errorf("time parse error: %s", terr)
	}

	scalarMetric := metrics.NewMetric("foo", "prog", metrics.Counter)
	d, _ := scalarMetric.GetDatum()
	d.Set(37, ts)
	r := FakeSocketWrite(metricToGraphite, scalarMetric)
	expected := []string{"prog.foo 37 1343124840\n"}
	diff := pretty.Compare(r, expected)
	if len(diff) > 0 {
		t.Errorf("String didn't match:\n%s", diff)
	}

	dimensionedMetric := metrics.NewMetric("bar", "prog", metrics.Gauge, "l")
	d, _ = dimensionedMetric.GetDatum("quux")
	d.Set(37, ts)
	d, _ = dimensionedMetric.GetDatum("snuh")
	d.Set(37, ts)
	r = FakeSocketWrite(metricToGraphite, dimensionedMetric)
	expected = []string{
		"prog.bar.l.quux 37 1343124840\n",
		"prog.bar.l.snuh 37 1343124840\n"}
	diff = pretty.Compare(r, expected)
	if len(diff) > 0 {
		t.Errorf("String didn't match:\n%s", diff)
	}
}

func TestMetricToStatsd(t *testing.T) {
	ts, terr := time.Parse("2006/01/02 15:04:05", "2012/07/24 10:14:00")
	if terr != nil {
		t.Errorf("time parse error: %s", terr)
	}

	scalarMetric := metrics.NewMetric("foo", "prog", metrics.Counter)
	d, _ := scalarMetric.GetDatum()
	d.Set(37, ts)
	r := FakeSocketWrite(metricToStatsd, scalarMetric)
	expected := []string{"prog.foo:37|c"}
	if !reflect.DeepEqual(expected, r) {
		t.Errorf("String didn't match:\n\texpected: %v\n\treceived: %v", expected, r)
	}

	dimensionedMetric := metrics.NewMetric("bar", "prog", metrics.Gauge, "l")
	d, _ = dimensionedMetric.GetDatum("quux")
	d.Set(37, ts)
	d, _ = dimensionedMetric.GetDatum("snuh")
	d.Set(42, ts)
	r = FakeSocketWrite(metricToStatsd, dimensionedMetric)
	expected = []string{
		"prog.bar.l.quux:37|g",
		"prog.bar.l.snuh:42|g"}
	if !reflect.DeepEqual(expected, r) {
		t.Errorf("String didn't match:\n\texpected: %v\n\treceived: %v", expected, r)
	}

	timingMetric := metrics.NewMetric("foo", "prog", metrics.Timer)
	d, _ = timingMetric.GetDatum()
	d.Set(37, ts)
	r = FakeSocketWrite(metricToStatsd, timingMetric)
	expected = []string{"prog.foo:37|ms"}
	if !reflect.DeepEqual(expected, r) {
		t.Errorf("String didn't match:\n\texpected: %v\n\treceived: %v", expected, r)
	}
}

func TestMetricToDogstatsd(t *testing.T) {
	ts, terr := time.Parse("2006/01/02 15:04:05", "2012/07/24 10:14:00")
	if terr != nil {
		t.Errorf("time parse error: %s", terr)
	}

	scalarMetric := metrics.NewMetric("foo", "prog", metrics.Counter)
	d, _ := scalarMetric.GetDatum()
	d.Set(37, ts)
	r := FakeSocketWrite(metricToDogstatsd, scalarMetric)
	expected := []string{"foo:37|c|#program:prog"}
	if !reflect.DeepEqual(expected, r) {
		t.Errorf("String didn't match:\n\texpected: %v\n\treceived: %v", expected, r)
	}

	dimensionedMetric := metrics.NewMetric("bar", "prog", metrics.Gauge, "l", "x")
	d, _ = dimensionedMetric.GetDatum("quux", "ham")
	d.Set(37, ts)
	d, _ = dimensionedMetric.GetDatum("snuh", "spam")
	d.Set(42, ts)
	r = FakeSocketWrite(metricToDogstatsd, dimensionedMetric)

	// Parse the metrics to pull out tags for comparison
	result_metrics := []string{}
	result_tags := []map[string]string{}
	for _, result := range r {
		split := strings.Split(result, "#")
		result_metrics = append(result_metrics, split[0])
		raw_tags := split[1]
		tags := map[string]string{}
		for _, tagval := range strings.Split(raw_tags, ",") {
			parts := strings.Split(tagval, ":")
			tags[parts[0]] = parts[1]
		}
		result_tags = append(result_tags, tags)
	}

	expected_metrics := []string{
		"bar:37|g|",
		"bar:42|g|"}
	if !reflect.DeepEqual(expected_metrics, result_metrics) {
		t.Errorf("Metric prefixes didn't match:\n\texpected: %v\n\treceived: %v", expected_metrics, result_metrics)
	}
	expected_tags := []map[string]string{
		{"l": "quux", "program": "prog", "x": "ham"},
		{"l": "snuh", "program": "prog", "x": "spam"},
	}
	if !reflect.DeepEqual(expected_tags, result_tags) {
		t.Errorf("Tags didn't match:\n\texpected: %v\n\treceived: %v", expected_tags, result_tags)
	}

	timingMetric := metrics.NewMetric("foo", "prog", metrics.Timer)
	d, _ = timingMetric.GetDatum()
	d.Set(37, ts)
	r = FakeSocketWrite(metricToDogstatsd, timingMetric)
	expected = []string{"foo:37|ms|#program:prog"}
	if !reflect.DeepEqual(expected, r) {
		t.Errorf("String didn't match:\n\texpected: %v\n\treceived: %v", expected, r)
	}
}
