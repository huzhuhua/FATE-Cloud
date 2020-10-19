package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type DeployComponent struct {
	Id               int64 `gorm:"type:bigint(12);column:id;primary_key;AUTO_INCREMENT"`
	FederatedId      int
	PartyId          int
	SiteName         string
	ProductType      int
	JobId            string
	FateVersion      string
	ComponentVersion string
	ComponentName    string
	Address          string
	StartTime        time.Time
	EndTime          time.Time
	Duration         int
	VersionIndex     int
	DeployStatus     int
	Status           int
	IsValid          int
	FinishTime       time.Time
	CreateTime       time.Time
	UpdateTime       time.Time
}

func GetDeployComponent(info DeployComponent) ([]*DeployComponent, error) {
	var deployComponentList []*DeployComponent
	Db := db
	if info.PartyId > 0 {
		Db = Db.Where("party_id = ?", info.PartyId)
	}
	if len(info.ComponentName) > 0 {
		Db = Db.Where("component_name = ?", info.ComponentName)
	}
	if info.ProductType >= 0 {
		Db = Db.Where("product_type = ?", info.ProductType)
	}
	if info.IsValid > 0 {
		Db = Db.Where("is_valid = ?", info.IsValid)
	}
	if info.DeployStatus > 0 {
		Db = Db.Where("deploy_status = ?", info.DeployStatus)
	}
	err := Db.Find(&deployComponentList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return deployComponentList, nil
}

func UpdateDeployComponent(info map[string]interface{}, condition DeployComponent) error {
	Db := db
	if condition.PartyId > 0 {
		Db = Db.Where("party_id = ?", condition.PartyId)
	}
	if len(condition.ComponentName) > 0 {
		Db = Db.Where("component_name = ?", condition.ComponentName)
	}
	if condition.ProductType > 0 {
		Db = Db.Where("product_type = ?", condition.ProductType)
	}
	if condition.IsValid >= 0 {
		Db = Db.Where("is_valid = ?", condition.IsValid)
	}
	if condition.DeployStatus >0 {
		Db = Db.Where("deploy_status = ?", condition.DeployStatus)
	}
	if err := Db.Model(&DeployComponent{}).Updates(info).Error; err != nil {
		return err
	}
	return nil
}

func AddDeployComponent(deployComponent *DeployComponent) error {
	if err := db.Create(&deployComponent).Error; err != nil {
		return err
	}
	return nil
}
