package models

import (
	"fmt"
	"github.com/YuZongYangHi/cloud-controller-manager/pkg/config"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

var db *gorm.DB

func RegisterDatabase(dbConfig config.CloudProviderDBConfig) error {
	var err error

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		dbConfig.User,
		dbConfig.Password,
		fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port),
		dbConfig.Name,
		"Asia%2fShanghai",
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		klog.Errorf("connection mysql fail!!!: %s", err.Error())
		return err
	}

	klog.Info("connection mysql successfully!")
	return nil
}

func Cursor() *gorm.DB {
	return db
}
