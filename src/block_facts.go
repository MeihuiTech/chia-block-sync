package main

import (
	"errors"
	"gorm.io/gorm"
	"math"
	"time"
)

const EpochBlocks = 4608
//https://github.com/Chia-Network/chia-blockchain/issues/2182
//The target difficulty is to have 4608 blocks per 24 hours. Since space is growing nearly 40% per week, it is accelerating the daily blocks by about 8% before the difficulty reset that happens each 4608 blocks.
const DailyBlocks = EpochBlocks * 1.08
const FirstBlockTimestamp = 1616162474
const SecondsPerBlock = (24 * 3600) / DailyBlocks


type ChiaTotalFarmerBlocks struct {
	ID            uint64 `gorm:"primaryKey;<-:false" json:"id"`
	FarmerAddress string `gorm:"type:varchar(256);not null;index:idx_tfb_farmer_address" json:"farmer_address"`
	BlockCount    uint64 `gorm:"type:bigint(20);not null;"`
}

type ChiaDailyFarmerBlocks struct {
	ID            uint64 `gorm:"primaryKey;<-:false" json:"id"`
	FarmerAddress string `gorm:"type:varchar(256);not null;index:idx_dfb_farmer_address" json:"farmer_address"`
	BlockCount    uint64 `gorm:"type:bigint(20);not null;"`
	Day           string `gorm:"type:date;not null;index:idx_dfb_day"`
}

type ChiaBlockSyncHeight struct {
	ID     uint64 `gorm:"primaryKey;<-:false" json:"id"`
	Height uint64 `gorm:"type:bigint(20);not null;default:0" json:"height"`
}

func IncreaseTotalBlock(farmerAddress string, db *gorm.DB) error {
	var totalBlock ChiaTotalFarmerBlocks
	r := db.Where("farmer_address = ?", farmerAddress).Take(&totalBlock)
	if r.Error == nil {
		r = db.Model(totalBlock).Update("block_count", gorm.Expr("block_count + ?", 1))
		return r.Error
	} else if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		r = db.Create(&ChiaTotalFarmerBlocks{
			BlockCount: 1,
			FarmerAddress: farmerAddress,
		})
		return r.Error
	} else {
		return r.Error
	}
}

// estimate timestamp base on block height
func HeightToTimestamp(height uint64) uint64 {
	return uint64(math.Round(SecondsPerBlock* float64(height))) + FirstBlockTimestamp
}

func IncreaseDailyBlock(farmerAddress string,timestamp uint64, db *gorm.DB) error {
	timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02") //设置时间戳 使用模板格式化为日期字符串

	var dailyBlock ChiaDailyFarmerBlocks
	r := db.Where("farmer_address = ? and day = ?", farmerAddress, timeStr).Take(&dailyBlock)
	if r.Error == nil {
		r = db.Model(dailyBlock).Update("block_count", gorm.Expr("block_count + ?", 1))
		return r.Error
	} else if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		r := db.Create(&ChiaDailyFarmerBlocks{
			BlockCount: 1,
			FarmerAddress: farmerAddress,
			Day:timeStr,
		})
		return r.Error
	} else {
		return r.Error
	}
}

func LogSyncHeight(height uint64, db *gorm.DB) error {
	blockHeight, err := GetSyncedHeight(db)
	if err != nil {
		return err
	}
	if blockHeight == nil {
		r := db.Create(&ChiaBlockSyncHeight{
			Height: height,
		})
		return r.Error
	} else {
		r := db.Model(blockHeight).Update("height", height)
		return r.Error
	}
}
