package collector

import (
	"context"
	"slices"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

type datastorePerfCollector struct {
	extraLabels []string

	scraper *scraper.VCenterScraper

	perfMetric *prometheus.Desc
}

func NewDatastorePerfCollector(scraper *scraper.VCenterScraper, cConf config.CollectorConfig) *datastorePerfCollector {
	labels := []string{"id", "name", "cluster", "datastore_kind", "datacenter"}
	extraLabels := cConf.DatastoreTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	perfLabels := append(slices.Clone(labels), "kind", "instance", "unit")

	return &datastorePerfCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		perfMetric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "perf_metric"),
			"Performance metric", perfLabels, nil),
	}
}

func (c *datastorePerfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.perfMetric
}

func (c *datastorePerfCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.scraper.Datastore.Enabled() || !c.scraper.DatastorePerf.Enabled() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), COLLECT_TIMEOUT)
	defer cancel()

	datastores, err := c.scraper.DB.GetAllDatastore(ctx)
	if err != nil && Logger != nil {
		Logger.Error("failed to get datastores", "err", err)
	}
	for _, datastore := range datastores {

		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, datastore.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{datastore.Self.ID(), datastore.Name, datastore.DatastoreCluster, datastore.Kind, datastore.Datacenter}
		labelValues = append(labelValues, extraLabelValues...)

		for metric := range c.scraper.MetricsDB.PopAllDatastoreMetricsIter(ctx, datastore.Self) {
			perfMetricLabelValues := append(slices.Clone(labelValues), metric.Name, metric.Instance, metric.Unit)
			ch <- prometheus.NewMetricWithTimestamp(metric.Timestamp, prometheus.MustNewConstMetric(
				c.perfMetric, prometheus.GaugeValue, metric.Value, perfMetricLabelValues...,
			))
		}
	}
}
