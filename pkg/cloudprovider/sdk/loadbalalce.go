package sdk

import (
	"errors"
	"github.com/YuZongYangHi/cloud-controller-manager/pkg/config"
	"github.com/YuZongYangHi/cloud-controller-manager/pkg/util/parsers"
	"k8s.io/klog/v2"
	"sync"
)

type serviceCache struct {
	mu             sync.RWMutex
	loadBalanceMap map[string]*LoadBalance
}

type LoadBalanceClient struct {
	LoadBalanceConfig *config.LoadBalanceConfig
	httpClient        *HTTPClient
	serviceCache      *serviceCache
}

func (c *serviceCache) GetByKey(key string) *LoadBalance {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loadBalanceMap[key]
}

func (c *serviceCache) delete(ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.loadBalanceMap, ip)
}

func (c *serviceCache) set(ip string, balance *LoadBalance) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadBalanceMap[ip] = balance
}

func (c *serviceCache) pop() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var lastKey string
	var ip string
	for key, _ := range c.loadBalanceMap {
		lastKey = key
	}
	if lastKey == "" {
		return "", errors.New("no available ip by cache")
	}
	v := c.loadBalanceMap[lastKey]
	ip = v.Ip
	delete(c.loadBalanceMap, lastKey)
	return ip, nil
}

func (c *LoadBalanceClient) Bind(name, namespace, ip string) error {
	var result *LoadBalanceMetadata
	m := &LoadBalance{
		Cluster:     c.LoadBalanceConfig.Region,
		Ip:          ip,
		Namespace:   namespace,
		ServiceName: name,
	}
	body := &PostOrPutParams{
		URL:         c.LoadBalanceConfig.LoadBalanceSet.Bind,
		Body:        m,
		Empowerment: &result,
	}
	err := c.httpClient.POST(body)
	if err != nil {
		return err
	}
	if result.Code != 200 {
		return errors.New(result.Message)
	}
	return nil
}

func (c *LoadBalanceClient) Unbind(name, namespace string) error {
	var result *LoadBalanceMetadata
	m := &LoadBalance{
		ServiceName: name,
		Namespace:   namespace,
	}
	body := &PostOrPutParams{
		URL:         c.LoadBalanceConfig.LoadBalanceSet.Released,
		Body:        m,
		Empowerment: &result,
	}
	err := c.httpClient.POST(body)
	if err != nil {
		return err
	}
	if result.Code != 200 {
		return errors.New(result.Message)
	}
	return nil
}

func (c *LoadBalanceClient) GetAvailableIp() (string, error) {
	ipList, err := c.List()
	if err == nil && len(*ipList) > 0 {
		return (*ipList)[0].Ip, nil
	}

	if err == nil && len(*ipList) == 0 {
		return "", errors.New("no available ip")
	}
	return c.serviceCache.pop()
}

func (c *LoadBalanceClient) List() (result *[]LoadBalance, err error) {
	var metadata *LoadBalanceMetadata
	params := &GetOrDeleteParams{
		URL:         c.LoadBalanceConfig.LoadBalanceSet.List,
		Params:      map[string]string{"status": "0"},
		Empowerment: &metadata,
	}
	err = c.httpClient.GET(params)
	if err != nil {
		return nil, err
	}

	err = parsers.JsonInterface(metadata.Data, &result)
	if err != nil {
		return nil, err
	}
	return
}

func (c *LoadBalanceClient) WaitForCacheSync() bool {
	fullData, err := c.List()
	if err != nil {
		klog.Errorf("sync loadbalance ip list fulldata error: %s", err.Error())
		return false
	}

	for _, data := range *fullData {
		c.serviceCache.set(data.Ip, &data)
	}

	klog.Infof("sync full loadbalance success, current cache count: %d", len(*fullData))

	return true
}

func NewLoadBalance(config *config.LoadBalanceConfig) *LoadBalanceClient {
	c := &LoadBalanceClient{
		LoadBalanceConfig: config,
		httpClient:        NewHTTPClient(),
		serviceCache:      &serviceCache{loadBalanceMap: make(map[string]*LoadBalance)},
	}
	return c
}
