package loadbalance

import "github.com/YuZongYangHi/cloud-controller-manager/cmd/cloud-provider-manager/models"

func Valid(m *models.LoadBalance) bool {
	if m.Ip == "" || m.ServiceName == "" || m.Namespace == "" {
		return false
	}
	return true
}
