package main

import (
	"context"
	"errors"
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
	code, err := v.Db.GetCode(&pb.Code{Id: uv.CodeId})
	if err != nil {
		return nil, err
	}
	uv.State = in.State
	uv.UsedAt = in.UsedAt
	if err := v.Db.TransUpdateUserVoucher(uv, code); err != nil {
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
