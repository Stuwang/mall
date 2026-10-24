package imdao

import "gorm.io/gorm"

type MemberDao struct {
	db *gorm.DB
}

func (d *MemberDao) WithTx(tx *gorm.DB) *MemberDao {
	return &MemberDao{db: defaultDB(tx)}
}
