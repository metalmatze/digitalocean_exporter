package collector

import (
	"context"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LoadBalancerCollector struct {
	logger  log.Logger
	client  *godo.Client
	timeout time.Duration
}

func NewLoadBalancerCollector(logger log.Logger, client *godo.Client, timeout time.Duration) *LoadBalancerCollector {
	return &LoadBalancerCollector{
		logger:  logger,
		client:  client,
		timeout: timeout,
	}
}

func (c *LoadBalancerCollector) Describe(ch chan<- *interface{}) {

}

func (c *LoadBalancerCollector) Collect(ch chan<- interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	lbs, _, err := c.client.LoadBalancers.List(ctx, nil)
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "can't list keys",
			"err", err,
		)
	}

	for range lbs {
		// TODO
	}
}
