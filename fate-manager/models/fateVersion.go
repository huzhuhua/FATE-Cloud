package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type FateVersion struct {
	Id           int64 `gorm:"type:bigint(12);column:id;primary_key;AUTO_INCREMENT"`
	FateVersion  string
	ProductType  int
	ChartVersion string
	VersionIndex int
	PullStatus   int
	CreateTime   time.Time
	UpdateTime   time.Time
}

func GetFateVersionList(fateVersion *FateVersion) ([]*FateVersion, error) {
	var fateVersionList []*FateVersion
	Db := db
	if fateVersion.ProductType > 0 {
		Db = Db.Where("product_type = ?", fateVersion.ProductType)
	}
	if fateVersion.VersionIndex > 0 {
		Db = Db.Where("version_index > ?", fateVersion.VersionIndex)
	}
	if len(fateVersion.FateVersion) > 0 {
		Db = Db.Where("fate_version = ?", fateVersion.FateVersion)
	}
	if fateVersion.PullStatus > 0 {
		Db = Db.Where("pull_status = ?", fateVersion.PullStatus)
	}
	if len(fateVersion.ChartVersion) >0{
		Db = Db.Where("chart_version = ?", fateVersion.ChartVersion)
	}
	err := Db.Group("fate_version").Find(&fateVersionList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return fateVersionList, nil
}

func AddFateVersion(fateVersion *FateVersion) error {
	if err := db.Create(&fateVersion).Error; err != nil {
		return err
	}
	return nil
}
