package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	pb "github.com/huyshop/header/voucher"
	"github.com/huyshop/voucher/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Voucher struct {
	Db    IDatabase
	cache *redis.Client
}

type IDatabase interface {
	InsertVoucher(req *pb.Voucher) (*pb.Voucher, error)
	GetVoucher(req *pb.Voucher) (*pb.Voucher, error)
	ListVoucher(req *pb.VoucherRequest) ([]*pb.Voucher, error)
	IsVoucherExist(req *pb.Voucher) (bool, error)
	UpdateVoucher(req *pb.Voucher) error
	DeleteVoucher(req *pb.Voucher) error
	CountVouchers(req *pb.VoucherRequest) (int64, error)

	InsertCode(req *pb.Code) error
	TransInsertCode(req *pb.Code) error
	TransInsertUserVoucher(req *pb.UserVoucher, code *pb.Code) error
	TransUpdateUserVoucher(uv *pb.UserVoucher) error
	GetCode(req *pb.Code) (*pb.Code, error)
	IsCodeExist(req *pb.Code) (bool, error)
	ListCode(req *pb.CodeRequest) ([]*pb.Code, error)
	CountCode(req *pb.CodeRequest) (int64, error)
	UpdateCode(req *pb.Code) error

	InsertUserVoucher(req *pb.UserVoucher) (*pb.UserVoucher, error)
	GetUserVoucher(req *pb.UserVoucher) (*pb.UserVoucher, error)
	ListUserVoucher(req *pb.UserVoucherRequest) ([]*pb.UserVoucher, error)
	UpdateUserVoucher(req *pb.UserVoucher) error
	DeleteUserVoucher(req *pb.UserVoucher) error
	CountUserVoucher(req *pb.UserVoucherRequest) (int64, error)
}

func NewRedisCache(addr, pw string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pw,
		DB:       db,
	})
	tick := time.NewTicker(10 * time.Minute)
	ctx := context.Background()
	go func(client *redis.Client) {
		for {
			select {
			case <-tick.C:
				if err := client.Ping(ctx).Err(); err != nil {
					panic(err)
				}
			}
		}
	}(client)
	return client
}

func NewVoucher(cf *Configs) (*Voucher, error) {
	dbase := &db.DB{}
	if err := dbase.ConnectDb(cf.DBPath, cf.DBName); err != nil {
		return nil, err
	}
	log.Println("Connect db successful")
	redisDb, _ := strconv.Atoi(config.RedisDb)
	rd := NewRedisCache(config.RedisAddr, config.RedisPassword, redisDb)
	log.Println("Connect redis successful")
	return &Voucher{
		Db:    dbase,
		cache: rd,
	}, nil
}

func startGRPCServe(port string, p *Voucher) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: 15 * time.Second,
		}),
	}
	serve := grpc.NewServer(opts...)
	pb.RegisterVoucherServiceServer(serve, p)
	reflection.Register(serve)
	return serve.Serve(listen)
}
