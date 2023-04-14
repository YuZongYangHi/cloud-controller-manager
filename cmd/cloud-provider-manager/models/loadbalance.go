package models

import (
	"errors"
	"time"
)

const TableNameLoadBalance = "loadbalances"

type LoadBalance struct {
	Id          int64     `json:"id"`
	Cluster     string    `json:"cluster"`
	Ip          string    `json:"ip"`
	Carriers    int       `json:"carriers"`
	Status      int       `json:"status"`
	Cidr        string    `json:"cidr"`
	Namespace   string    `json:"namespace"`
	ServiceName string    `json:"serviceName"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (*LoadBalance) TableName() string {
	return TableNameLoadBalance
}

type loadBalanceModel struct{}

func (c *loadBalanceModel) List(values map[string]interface{}) (result *[]LoadBalance) {
	db.Where(values).Find(&result)
	return
}

func (c *loadBalanceModel) GetByIp(ip string) (*LoadBalance, error) {
	var result *LoadBalance
	err := db.Where("ip = ?", ip).First(&result).Error
	return result, err
}

func (c *loadBalanceModel) GetByService(name, namespace string) (*LoadBalance, error) {
	var result *LoadBalance
	err := db.Where("service_name = ? AND namespace = ?", name, namespace).First(&result).Error
	return result, err
}

func (c *loadBalanceModel) Bind(m *LoadBalance) (*LoadBalance, error) {
	obj, err := c.GetByIp(m.Ip)
	if err != nil {
		return nil, err
	}

	if obj.Status == 1 {
		return nil, errors.New("ip has been bound")
	}

	obj.UpdatedAt = time.Now()
	obj.Status = 1
	obj.Namespace = m.Namespace
	obj.ServiceName = m.ServiceName
	return obj, db.Save(&obj).Error
}

func (c *loadBalanceModel) Released(m *LoadBalance) error {
	obj, err := c.GetByService(m.ServiceName, m.Namespace)
	if err != nil {
		return err
	}
	obj.Namespace = ""
	obj.ServiceName = ""
	obj.Status = 0
	obj.UpdatedAt = time.Now()
	return db.Save(&obj).Error
}
