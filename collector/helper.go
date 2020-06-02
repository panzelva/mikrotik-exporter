package collector

import (
	"math"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func metricStringCleanup(in string) string {
	return strings.Replace(in, "-", "_", -1)
}

func descriptionForPropertyName(prefix, property string, labelNames []string) *prometheus.Desc {
	return descriptionForPropertyNameHelpText(prefix, property, labelNames, property)
}

func descriptionForPropertyNameHelpText(prefix, property string, labelNames []string, helpText string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefix, metricStringCleanup(property)),
		helpText,
		labelNames,
		nil,
	)
}

func description(prefix, name, helpText string, labelNames []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefix, name),
		helpText,
		labelNames,
		nil,
	)
}

func splitStringToFloats(metric string) (float64, float64, error) {
	strs := strings.Split(metric, ",")
	if len(strs) == 0 {
		return 0, 0, nil
	}
	m1, err := strconv.ParseFloat(strs[0], 64)
	if err != nil {
		return math.NaN(), math.NaN(), err
	}
	m2, err := strconv.ParseFloat(strs[1], 64)
	if err != nil {
		return math.NaN(), math.NaN(), err
	}
	return m1, m2, nil
}

func parseStringToFloat64(value, device, property, errorMsg string) (float64, error) {
	parsedValue, err := strconv.ParseFloat(value, 64)

	if err != nil {
		if errorMsg != "" {
			log.WithFields(log.Fields{"device": device, "property": property, "value": value, "error": err}).Error(errorMsg)
		}
		return math.NaN(), err
	}

	return parsedValue, nil
}
