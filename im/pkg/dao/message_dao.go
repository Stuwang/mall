package imdao

import "gorm.io/gorm"

type MessageDao struct {
	db *gorm.DB
}

func (d *MessageDao) WithTx(tx *gorm.DB) *MessageDao {
	return &MessageDao{db: defaultDB(tx)}
}
