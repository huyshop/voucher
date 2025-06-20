package db

import (
	"log"

	pb "github.com/huyshop/header/voucher"
	"xorm.io/xorm"
)

const (
	tblVoucher     = "voucher"
	tblUserVoucher = "user_voucher"
	tblCode        = "code"
)

func createTable(model interface{}, tblName string, engine *xorm.Engine) error {
	b, err := engine.IsTableExist(model)
	if err != nil {
		return err
	}
	log.Print(b, " ", tblName)
	if b {
		if err = engine.Sync2(model); err != nil {
			return err
		}
		return nil
	}
	if !b {
		if err := engine.CreateTables(model); err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

func (d *DB) CreateDb() error {
	if err := createTable(&pb.Voucher{}, tblVoucher, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.UserVoucher{}, tblUserVoucher, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.Code{}, tblCode, d.engine); err != nil {
		return err
	}
	return nil
}
