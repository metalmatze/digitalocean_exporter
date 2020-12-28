package collector

import (
	"context"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// BalanceCollector collects metrics about the account.
type BalanceCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	MonthToDateBalance *prometheus.Desc
	AccountBalance     *prometheus.Desc
	MonthToDateUsage   *prometheus.Desc
	BalanceGeneratedAt *prometheus.Desc
}

// NewBalanceCollector returns a new BalanceCollector.
func NewBalanceCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *BalanceCollector {
	errors.WithLabelValues("balance").Add(0)

	return &BalanceCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		MonthToDateBalance: prometheus.NewDesc(
			"digitalocean_month_to_date_balance",
			"Balance as of the digitalocean_balance_generated_at time",
			nil, nil,
		),
		AccountBalance: prometheus.NewDesc(
			"digitalocean_account_balance",
			"Current balance of your most recent billing activity",
			nil, nil,
		),
		MonthToDateUsage: prometheus.NewDesc(
			"digitalocean_month_to_date_usage",
			"Amount used in the current billing period as of the digitalocean_balance_generated_at time",
			nil, nil,
		),
		BalanceGeneratedAt: prometheus.NewDesc(
			"digitalocean_balance_generated_at",
			"The time at which balances were most recently generated",
			nil, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *BalanceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.MonthToDateBalance
	ch <- c.AccountBalance
	ch <- c.MonthToDateUsage
	ch <- c.BalanceGeneratedAt
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *BalanceCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	bal, _, err := c.client.Balance.Get(ctx)
	if err != nil {
		c.errors.WithLabelValues("balance").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't get balance",
			"err", err,
		)
		return
	}

	monthToDateBalance, err := strconv.ParseFloat(bal.MonthToDateBalance, 64)
	if err != nil {
		c.errors.WithLabelValues("balance").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't parse MonthToDateBalance",
			"err", err,
		)
		monthToDateBalance = -1
	}
	ch <- prometheus.MustNewConstMetric(
		c.MonthToDateBalance,
		prometheus.GaugeValue,
		monthToDateBalance,
	)
	accountBalance, err := strconv.ParseFloat(bal.AccountBalance, 64)
	if err != nil {
		c.errors.WithLabelValues("balance").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't parse AccountBalance",
			"err", err,
		)
		accountBalance = -1
	}
	ch <- prometheus.MustNewConstMetric(
		c.AccountBalance,
		prometheus.GaugeValue,
		accountBalance,
	)
	monthToDateUsage, err := strconv.ParseFloat(bal.MonthToDateUsage, 64)
	if err != nil {
		c.errors.WithLabelValues("balance").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't parse MonthToDateUsage",
			"err", err,
		)
		monthToDateUsage = -1
	}
	ch <- prometheus.MustNewConstMetric(
		c.MonthToDateUsage,
		prometheus.GaugeValue,
		monthToDateUsage,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BalanceGeneratedAt,
		prometheus.GaugeValue,
		float64(bal.GeneratedAt.Unix()),
	)
}
