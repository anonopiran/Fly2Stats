package v2ray

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DirectionType string

const (
	Downlink DirectionType = "downlink"
	Uplink   DirectionType = "uplink"
)

type UserStatType struct {
	Username  string
	Time      int64
	Direction DirectionType
	Value     int64
}
type IStat interface {
	GetName() string
	GetValue() int64
}
type IStatsRequest interface {
	Descriptor() ([]byte, []int)
	GetPattern() string
	GetPatterns() []string
	GetRegexp() bool
	GetReset_() bool
	ProtoMessage()
	ProtoReflect() protoreflect.Message
	Reset()
	String() string
}
type IStatsResponse interface {
	Descriptor() ([]byte, []int)
	GetStat() []IStat
	ProtoMessage()
	ProtoReflect() protoreflect.Message
	Reset()
	String() string
}

type StatsServiceClient interface {
	QueryStats(ctx context.Context, in IStatsRequest, opts ...grpc.CallOption) ([]IStat, error)
}
