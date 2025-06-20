package main

import (
	"context"
	"errors"
	"time"

	pb "github.com/huyshop/header/voucher"
	"github.com/huyshop/voucher/utils"
)

func (v *Voucher) AssignVoucherToUser(ctx context.Context, in *pb.UserVoucher) (*pb.UserVoucher, error) {
	if in.GetUserId() == "" || in.GetVoucherId() == "" {
		return nil, errors.New(utils.E_invalid_input)
	}
	in.Id = utils.MakeUserVoucherId()
	in.State = pb.UserVoucher_got.String()
	uv, err := v.Db.InsertUserVoucher(in)
	if err != nil {
		return nil, err
	}
	return uv, nil
}

func (v *Voucher) UseUserVoucher(ctx context.Context, in *pb.UserVoucher) (*pb.UserVoucher, error) {
	if in.Id == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	uv, err := v.Db.GetUserVoucher(in)
	if err != nil {
		return nil, err
	}
	uv.State = pb.UserVoucher_used.String()
	uv.UsedAt = time.Now().Unix()
	if err := v.Db.UpdateUserVoucher(uv); err != nil {
		return nil, err
	}
	return uv, nil
}

func (v *Voucher) ListUserVouchers(ctx context.Context, in *pb.UserVoucherRequest) (*pb.UserVouchers, error) {
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
	uv, err := v.Db.GetUserVoucher(in)
	if err != nil {
		return nil, err
	}
	return uv, nil
}
