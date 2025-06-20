package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/voucher"
	"github.com/huyshop/voucher/utils"
)

func (v *Voucher) CreateVoucher(ctx context.Context, in *pb.Voucher) (*common.Empty, error) {
	if in.GetName() == "" {
		return nil, errors.New(utils.E_invalid_input)
	}
	if in.GetDiscount() <= 0 && in.GetPointExchange() <= 0 {
		return nil, errors.New(utils.E_invalid_discount)
	}
	in.Id = utils.MakeVoucherId()
	in.CreatedAt = time.Now().Unix()
	_, err := v.Db.InsertVoucher(in)
	if err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (v *Voucher) UpdateVoucher(ctx context.Context, in *pb.Voucher) (*common.Empty, error) {
	if in.GetId() == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	in.UpdatedAt = time.Now().Unix()
	if err := v.Db.UpdateVoucher(in); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (v *Voucher) DeleteVoucher(ctx context.Context, in *pb.Voucher) (*common.Empty, error) {
	if in.GetId() == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	if err := v.Db.DeleteVoucher(in); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (v *Voucher) GetVoucher(ctx context.Context, in *pb.Voucher) (*pb.Voucher, error) {
	return v.Db.GetVoucher(in)
}

func (v *Voucher) ListVouchers(ctx context.Context, in *pb.VoucherRequest) (*pb.Vouchers, error) {
	log.Println("in ", in)
	list, err := v.Db.ListVoucher(in)
	if err != nil {
		return nil, err
	}
	count, err := v.Db.CountVouchers(in)
	if err != nil {
		return nil, err
	}
	return &pb.Vouchers{Vouchers: list, Total: int32(count)}, nil
}
