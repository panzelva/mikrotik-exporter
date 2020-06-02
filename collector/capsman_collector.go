package collector

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
)

type capsManCollector struct {
	props        []string
	descriptions map[string]*prometheus.Desc
}

func newCapsCollector() routerOSCollector {
	c := &capsManCollector{}
	c.init()
	return c
}

func (c *capsManCollector) init() {
	c.props = []string{"interface", "mac-address", "rx-signal", "packets", "bytes"}
	labelNames := []string{"name", "address", "interface", "mac_address"}
	c.descriptions = make(map[string]*prometheus.Desc)

	for _, p := range []string{"interface", "mac-address", "rx-signal"} {
		c.descriptions[p] = descriptionForPropertyName("caps_man", p, labelNames)
	}

	for _, p := range []string{"bytes", "packets"} {
		c.descriptions["tx_"+p] = descriptionForPropertyName("caps_man", "tx_"+p, labelNames)
		c.descriptions["rx_"+p] = descriptionForPropertyName("caps_man", "rx_"+p, labelNames)
	}
}

func (c *capsManCollector) describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *capsManCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectForStat(re, ctx)
	}

	return nil
}

func (c *capsManCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/caps-man/registration-table/print", "=.proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching wlan station metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *capsManCollector) collectForStat(re *proto.Sentence, ctx *collectorContext) {
	iface := re.Map["interface"]
	mac := re.Map["mac-address"]

	for _, p := range []string{"rx-signal"} {
		c.collectMetricForProperty(p, iface, mac, re, ctx)
	}
	for _, p := range []string{"bytes", "packets"} {
		c.collectMetricForTXRXCounters(p, iface, mac, re, ctx)
	}
}

func (c *capsManCollector) collectMetricForProperty(property, iface, mac string, re *proto.Sentence, ctx *collectorContext) {
	if re.Map[property] == "" {
		return
	}
	v, err := strconv.ParseFloat(re.Map[property], 64)
	if err != nil {
		log.WithFields(log.Fields{
			"device":   ctx.device.Name,
			"property": property,
			"value":    re.Map[property],
			"error":    err,
		}).Error("error parsing wlan station metric value")
		return
	}

	desc := c.descriptions[property]
	ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, ctx.device.Name, ctx.device.Address, iface, mac)
}

func (c *capsManCollector) collectMetricForTXRXCounters(property, iface, mac string, re *proto.Sentence, ctx *collectorContext) {
	tx, rx, err := splitStringToFloats(re.Map[property])
	if err != nil {
		log.WithFields(log.Fields{
			"device":   ctx.device.Name,
			"property": property,
			"value":    re.Map[property],
			"error":    err,
		}).Error("error parsing wlan station metric value")
		return
	}

	descTx := c.descriptions["tx_"+property]
	descRx := c.descriptions["rx_"+property]
	ctx.ch <- prometheus.MustNewConstMetric(descTx, prometheus.CounterValue, tx, ctx.device.Name, ctx.device.Address, iface, mac)
	ctx.ch <- prometheus.MustNewConstMetric(descRx, prometheus.CounterValue, rx, ctx.device.Name, ctx.device.Address, iface, mac)
}
