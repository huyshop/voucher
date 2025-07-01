package db

import (
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	pb "github.com/huyshop/header/voucher"
	"github.com/huyshop/voucher/utils"
	"xorm.io/xorm"
)

type DB struct {
	engine *xorm.Engine
}

func (d *DB) ConnectDb(sqlPath, dbName string) error {
	sqlConnStr := fmt.Sprintf("%s/%s", sqlPath, dbName)
	engine, err := xorm.NewEngine("mysql", sqlConnStr)
	if err != nil {
		return err
	}
	tickPingSql := time.NewTicker(15 * time.Minute)
	go func() {
		for {
			select {
			case <-tickPingSql.C:
				if err := engine.Ping(); err != nil {
					log.Print("sql can not ping")
				}
			}
		}
	}()
	d.engine = engine
	d.engine.ShowSQL(false)
	return err
}

// voucher
func (d *DB) InsertVoucher(req *pb.Voucher) (*pb.Voucher, error) {
	count, err := d.engine.Insert(req)
	if err != nil {
		return nil, err
	}
	if count < 1 {
		return nil, errors.New(utils.E_can_not_insert)
	}
	return req, nil
}

func (d *DB) GetVoucher(req *pb.Voucher) (*pb.Voucher, error) {
	voucher := &pb.Voucher{Id: req.Id}
	b, err := d.engine.Get(voucher)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, errors.New(utils.E_not_found_voucher)
	}
	return voucher, nil
}

func (d *DB) listVoucherQuery(req *pb.VoucherRequest) *xorm.Session {
	ss := d.engine.Table("voucher")
	if req.GetId() != "" {
		ss.And("id = ?", req.GetId())
	}
	if req.GetName() != "" {
		ss.And("name LIKE ?", "%"+req.GetName()+"%")
	}
	if req.GetPartnerId() != "" {
		ss.And("partner_id = ?", req.GetPartnerId())
	}
	if req.GetState() != "" {
		ss.And("state = ?", req.GetState())
	}
	if req.GetType() != "" {
		ss.And("type = ?", req.GetType())
	}
	return ss
}

func (d *DB) ListVoucher(req *pb.VoucherRequest) ([]*pb.Voucher, error) {
	voucher := []*pb.Voucher{}
	ss := d.listVoucherQuery(req)
	err := ss.Desc("created_at").Find(&voucher)
	if err != nil {
		log.Println("get list:", err)
		return nil, err
	}
	return voucher, nil
}

func (d *DB) IsVoucherExist(req *pb.Voucher) (bool, error) {
	b, err := d.engine.Exist(&pb.Voucher{Id: req.Id})
	if err != nil {
		return false, err
	}
	return b, err
}

func (d *DB) UpdateVoucher(req *pb.Voucher) error {
	count, err := d.engine.Update(req, &pb.Voucher{Id: req.Id})
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New(utils.E_can_not_update)
	}
	return nil
}

func (d *DB) DeleteVoucher(req *pb.Voucher) error {
	count, err := d.engine.Delete(req)
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New(utils.E_can_not_delete)
	}
	return nil
}

func (d *DB) CountVouchers(req *pb.VoucherRequest) (int64, error) {
	ss := d.listVoucherQuery(req)
	return ss.Count()
}

// code
func (d *DB) InsertCode(req *pb.Code) error {
	count, err := d.engine.Insert(req)
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New(utils.E_not_found)
	}
	return nil
}

func (d *DB) TransInsertCode(req *pb.Code) error {
	ss := d.engine.NewSession()
	defer ss.Close()
	if err := ss.Begin(); err != nil {
		return err
	}
	voucher := &pb.Voucher{Id: req.VoucherId}
	b, err := ss.Get(voucher)
	if err != nil {
		ss.Rollback()
		return err
	}
	if !b {
		ss.Rollback()
		return errors.New(utils.E_not_found_voucher)
	}
	if voucher.RemainingQuantity < 1 {
		ss.Rollback()
		return errors.New(utils.E_can_not_insert)
	}
	_, err = ss.Insert(req)
	if err != nil {
		ss.Rollback()
		return err
	}
	remainningQuantity := voucher.RemainingQuantity - 1
	_, err = ss.Cols("remaining_quantity", "updated_at").Update(&pb.Voucher{RemainingQuantity: remainningQuantity, UpdatedAt: time.Now().Unix()}, &pb.Voucher{Id: req.VoucherId})
	if err != nil {
		ss.Rollback()
		return err
	}
	if err := ss.Commit(); err != nil {
		ss.Rollback()
		return err
	}
	return nil
}

func (d *DB) TransInsertUserVoucher(req *pb.UserVoucher, code *pb.Code) error {
	ss := d.engine.NewSession()
	defer ss.Close()
	if err := ss.Begin(); err != nil {
		return err
	}
	voucher := &pb.Voucher{Id: req.VoucherId}
	b, err := ss.Get(voucher)
	if err != nil {
		ss.Rollback()
		return err
	}
	if !b {
		ss.Rollback()
		return errors.New(utils.E_not_found_voucher)
	}
	if voucher.RemainingQuantity < 1 {
		ss.Rollback()
		return errors.New(utils.E_can_not_insert)
	}
	_, err = ss.Insert(req)
	if err != nil {
		ss.Rollback()
		return err
	}
	code.State = pb.Code_got.String()
	code.CreatedAt = time.Now().Unix()
	_, err = ss.Insert(code)
	if err != nil {
		ss.Rollback()
		return err
	}
	remainningQuantity := voucher.RemainingQuantity - 1
	_, err = ss.Cols("remaining_quantity", "updated_at").Update(&pb.Voucher{RemainingQuantity: remainningQuantity, UpdatedAt: time.Now().Unix()}, &pb.Voucher{Id: req.VoucherId})
	if err != nil {
		ss.Rollback()
		return err
	}
	if err := ss.Commit(); err != nil {
		ss.Rollback()
		return err
	}
	return nil
}

func (d *DB) TransUpdateUserVoucher(uv *pb.UserVoucher, code *pb.Code) error {
	sess := d.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	count, err := sess.Update(uv, &pb.UserVoucher{
		UserId:    uv.GetUserId(),
		VoucherId: uv.GetVoucherId(),
		CodeId:    code.GetId(),
	})
	if err != nil {
		log.Print(err)
		sess.Rollback()
		return err
	}
	if count == 0 {
		sess.Rollback()
		return errors.New(utils.E_can_not_update)
	}
	count, err = sess.Update(code, &pb.Code{
		Code:      code.GetCode(),
		VoucherId: code.GetVoucherId(),
		Id:        code.GetId(),
	})
	if err != nil {
		log.Print(err)
		sess.Rollback()
		return err
	}
	if count == 0 {
		sess.Rollback()
		return errors.New(utils.E_can_not_update)
	}
	if err := sess.Commit(); err != nil {
		sess.Rollback()
		return err
	}
	return nil
}

func (d *DB) GetCode(req *pb.Code) (*pb.Code, error) {
	code := &pb.Code{Id: req.Id}
	b, err := d.engine.Get(code)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, errors.New(utils.E_not_found)
	}
	return code, nil
}

func (d *DB) IsCodeExist(req *pb.Code) (bool, error) {
	b, err := d.engine.Exist(&pb.Code{VoucherId: req.VoucherId, Code: req.Code})
	if err != nil {
		return false, err
	}
	return b, err
}

// code
func (d *DB) listCodeQuery(req *pb.CodeRequest) *xorm.Session {
	ss := d.engine.Table(tblCode)
	if req.GetCode() != "" {
		ss.And("code = ?", req.GetCode())
	}
	if req.VoucherId != "" {
		ss.And("voucher_id = ?", req.VoucherId)
	}
	if req.State != "" {
		ss.And("state = ?", req.State)
	}
	return ss
}

func (d *DB) ListCode(req *pb.CodeRequest) ([]*pb.Code, error) {
	code := []*pb.Code{}
	ss := d.listCodeQuery(req)
	err := ss.Desc(".created_at").Find(&code)
	if err != nil {
		log.Println("get list:", err)
		return nil, err
	}
	return code, nil
}

func (d *DB) CountCode(req *pb.CodeRequest) (int64, error) {
	ss := d.listCodeQuery(req)
	return ss.Count()
}

func (d *DB) UpdateCode(req *pb.Code) error {
	count, err := d.engine.Update(req, &pb.Code{Id: req.Id})
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New(utils.E_can_not_update)
	}
	return nil
}

// user_voucher
func (d *DB) InsertUserVoucher(req *pb.UserVoucher) (*pb.UserVoucher, error) {
	count, err := d.engine.Insert(req)
	if err != nil {
		return nil, err
	}
	if count < 1 {
		return nil, errors.New(utils.E_can_not_insert)
	}
	return req, nil
}

func (d *DB) GetUserVoucher(req *pb.UserVoucher) (*pb.UserVoucher, error) {
	uv := &pb.UserVoucher{Id: req.Id}
	b, err := d.engine.Get(uv)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, errors.New(utils.E_not_found_voucher)
	}
	return uv, nil
}

func (d *DB) listUserVoucherQuery(req *pb.UserVoucherRequest) *xorm.Session {
	ss := d.engine.Table("user_voucher")
	if len(req.GetIds()) > 0 {
		ss.In("id", req.GetIds())
	} else if req.GetId() != "" {
		ss.And("id = ?", req.GetId())
	}
	if req.GetType() != "" {
		ss.Join("INNER", "voucher", "user_voucher.voucher_id = voucher.id")
		ss.And("voucher.type = ?", req.GetType())
	}
	if req.GetState() != "" {
		ss.And("state = ?", req.GetState())
	}
	if req.GetUserId() != "" {
		ss.And("user_id = ?", req.GetUserId())
	}
	if req.GetVoucherId() != "" {
		ss.And("voucher_id = ?", req.GetVoucherId())
	}
	return ss
}

func (d *DB) ListUserVoucher(req *pb.UserVoucherRequest) ([]*pb.UserVoucher, error) {
	uv := []*pb.UserVoucher{}
	ss := d.listUserVoucherQuery(req)
	err := ss.Desc("user_voucher.created_at").Find(&uv)
	if err != nil {
		log.Println("get list:", err)
		return nil, err
	}
	return uv, nil
}

func (d *DB) UpdateUserVoucher(req *pb.UserVoucher) error {
	count, err := d.engine.Update(req, &pb.UserVoucher{Id: req.Id})
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New(utils.E_can_not_update)
	}
	return nil
}

func (d *DB) DeleteUserVoucher(req *pb.UserVoucher) error {
	count, err := d.engine.Delete(req)
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New(utils.E_can_not_delete)
	}
	return nil
}

func (d *DB) CountUserVoucher(req *pb.UserVoucherRequest) (int64, error) {
	ss := d.listUserVoucherQuery(req)
	return ss.Count()
}
