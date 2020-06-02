package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
)

type ipsecActivePeersCollector struct {
	props        []string
	descriptions map[string]*prometheus.Desc
}

func newIpsecActivePeersCollector() routerOSCollector {
	c := &ipsecActivePeersCollector{}
	c.init()
	return c
}

func (c *ipsecActivePeersCollector) init() {
	c.props = []string{"ph2-total", "uptime", "remote-address", "rx-bytes", "tx-bytes", "rx-packets", "tx-packets"}
	labelNames := []string{"name", "address", "remoteaddress"}
	c.descriptions = make(map[string]*prometheus.Desc)

	for _, property := range c.props {
		c.descriptions[property] = descriptionForPropertyName("ipsec_active_peers", property, labelNames)
	}
}

func (c *ipsecActivePeersCollector) describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *ipsecActivePeersCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectForStat(re, ctx)
	}

	return nil
}

func (c *ipsecActivePeersCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/ip/ipsec/active-peers/print", "=.proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching interface metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *ipsecActivePeersCollector) collectForStat(re *proto.Sentence, ctx *collectorContext) {
	for _, property := range c.props {
		c.collectMetricForProperty(property, re, ctx)
	}
}

func (c *ipsecActivePeersCollector) collectMetricForProperty(property string, re *proto.Sentence, ctx *collectorContext) {
	desc := c.descriptions[property]
	value := re.Map[property]
	remoteAddress := re.Map["remote-address"]

	switch property {
	case "tx-bytes", "rx-packets", "tx-packets", "rx-bytes":
		parsedValue, err := parseStringToFloat64(value, ctx.device.Name, property, "error parsing ipsec active peers value")
		if err != nil {
			return
		}

		// TODO counter value was collected before with the same name and label values
		ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, parsedValue, ctx.device.Name, ctx.device.Address, remoteAddress)
	case "ph2-total":
		parsedValue, err := parseStringToFloat64(value, ctx.device.Name, property, "")
		if err != nil {
			ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 0, ctx.device.Name, ctx.device.Address, remoteAddress)
			return
		}

		ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, parsedValue, ctx.device.Name, ctx.device.Address, remoteAddress)
	case "uptime":
		parsedValue, err := parseUptime(value)
		if err != nil {
			return
		}

		ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, parsedValue, ctx.device.Name, ctx.device.Address, remoteAddress)
	}
}
