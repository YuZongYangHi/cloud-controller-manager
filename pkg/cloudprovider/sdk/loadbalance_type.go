package sdk

type LoadBalanceMetadata struct {
	Data    interface{} `json:"data"`
	Code    int64       `json:"code"`
	Message string      `json:"message"`
}
type LoadBalance struct {
	Cluster     string `json:"cluster"`
	Ip          string `json:"ip"`
	Carriers    int    `json:"carriers"`
	Status      int    `json:"status"`
	Cidr        string `json:"cidr"`
	Namespace   string `json:"namespace"`
	ServiceName string `json:"serviceName"`
}
