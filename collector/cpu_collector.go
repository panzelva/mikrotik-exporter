package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
)

type cpuCollector struct {
	props        []string
	descriptions map[string]*prometheus.Desc
}

func newCPUCollector() routerOSCollector {
	c := &cpuCollector{}
	c.init()
	return c
}

func (c *cpuCollector) init() {
	c.props = []string{"cpu", "load", "irq", "disk"}
	labelNames := []string{"name", "address"}
	c.descriptions = make(map[string]*prometheus.Desc)

	for _, p := range c.props {
		c.descriptions[p] = descriptionForPropertyName("cpu", p, labelNames)
	}
}

func (c *cpuCollector) describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *cpuCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectForStat(re, ctx)
	}

	return nil
}

func (c *cpuCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/system/resource/cpu/print", "=.proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching cpu metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *cpuCollector) collectForStat(re *proto.Sentence, ctx *collectorContext) {
	for _, property := range c.props {
		c.collectMetricForProperty(property, re, ctx)
	}
}

func (c *cpuCollector) collectMetricForProperty(property string, re *proto.Sentence, ctx *collectorContext) {
	desc := c.descriptions[property]
	value := re.Map[property]
	parsedValue, err := parseStringToFloat64(value, ctx.device.Name, property, "")
	if err != nil {
		return
	}

	ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, parsedValue, ctx.device.Name, ctx.device.Address)
}
