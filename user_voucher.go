package main

import (
	"context"
	"errors"
	"log"
	"time"

	pb "github.com/huyshop/header/voucher"
	"github.com/huyshop/voucher/utils"
)

func (v *Voucher) CreateUserVoucher(ctx context.Context, in *pb.UserVoucher) (*pb.UserVoucher, error) {
	if in.UserId == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	if in.VoucherId == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	voucher, err := v.Db.GetVoucher(&pb.Voucher{Id: in.VoucherId})
	if err != nil {
		return nil, err
	}
	if voucher.RemainingQuantity == 0 {
		return nil, errors.New(utils.E_voucher_quantity_is_out)
	}
	if voucher.GetType() == pb.Voucher_free.String() {
		// user chỉ đc 1 code
		uv, err := v.Db.GetUserVoucher(&pb.UserVoucher{UserId: in.UserId, VoucherId: in.VoucherId})
		if err != nil {
			log.Println("err: ", err)
			return nil, err
		}
		if uv != nil {
			return nil, errors.New(utils.E_voucher_quantity_is_out)
		}
	}
	code := utils.MakeCode()
	codedata := &pb.Code{
		Id:        utils.MakeCodeId(),
		Code:      code,
		VoucherId: in.VoucherId,
		State:     pb.Code_got.String(),
		CreatedAt: time.Now().Unix(),
	}
	in.Id = utils.MakeUserVoucherId()
	in.CodeId = codedata.Id
	in.State = pb.UserVoucher_got.String()
	if err := v.Db.TransInsertUserVoucher(in, codedata); err != nil {
		return nil, err
	}
	return in, nil
}

func (v *Voucher) UpdateUserVoucher(ctx context.Context, in *pb.UserVoucher) (*pb.UserVoucher, error) {
	if in.UserId == "" || in.CodeId == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	uv, err := v.Db.GetUserVoucher(&pb.UserVoucher{
		UserId: in.UserId,
		CodeId: in.CodeId,
	})
	if err != nil {
		return nil, err
	}
	uv.State = in.State
	if err := v.Db.TransUpdateUserVoucher(uv); err != nil {
		return nil, err
	}
	return uv, nil
}

func (v *Voucher) ListUserVouchers(ctx context.Context, in *pb.UserVoucherRequest) (*pb.UserVouchers, error) {
	log.Println("ListUserVouchers", in)
	list, err := v.Db.ListUserVoucher(in)
	if err != nil {
		return nil, err
	}
	count, err := v.Db.CountUserVoucher(in)
	if err != nil {
		return nil, err
	}
	return &pb.UserVouchers{UserVouchers: list, Total: int32(count)}, nil
}

func (v *Voucher) GetUserVoucher(ctx context.Context, in *pb.UserVoucher) (*pb.UserVoucher, error) {

	log.Println("GetUserVoucher", in)
	uv, err := v.Db.GetUserVoucher(in)
	if err != nil {
		return nil, err
	}
	return uv, nil
}
