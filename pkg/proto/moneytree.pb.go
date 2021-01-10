// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: proto/moneytree.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type GetCandlesRequest_Duration int32

const (
	GetCandlesRequest_ONE_MINUTE        GetCandlesRequest_Duration = 0
	GetCandlesRequest_FIVE_MINUTES      GetCandlesRequest_Duration = 1
	GetCandlesRequest_FIFTEEN_MINUTES   GetCandlesRequest_Duration = 2
	GetCandlesRequest_ONE_HOUR          GetCandlesRequest_Duration = 3
	GetCandlesRequest_TWELVE_HOURS      GetCandlesRequest_Duration = 4
	GetCandlesRequest_TWENTY_FOUR_HOURS GetCandlesRequest_Duration = 5
)

// Enum value maps for GetCandlesRequest_Duration.
var (
	GetCandlesRequest_Duration_name = map[int32]string{
		0: "ONE_MINUTE",
		1: "FIVE_MINUTES",
		2: "FIFTEEN_MINUTES",
		3: "ONE_HOUR",
		4: "TWELVE_HOURS",
		5: "TWENTY_FOUR_HOURS",
	}
	GetCandlesRequest_Duration_value = map[string]int32{
		"ONE_MINUTE":        0,
		"FIVE_MINUTES":      1,
		"FIFTEEN_MINUTES":   2,
		"ONE_HOUR":          3,
		"TWELVE_HOURS":      4,
		"TWENTY_FOUR_HOURS": 5,
	}
)

func (x GetCandlesRequest_Duration) Enum() *GetCandlesRequest_Duration {
	p := new(GetCandlesRequest_Duration)
	*p = x
	return p
}

func (x GetCandlesRequest_Duration) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GetCandlesRequest_Duration) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_moneytree_proto_enumTypes[0].Descriptor()
}

func (GetCandlesRequest_Duration) Type() protoreflect.EnumType {
	return &file_proto_moneytree_proto_enumTypes[0]
}

func (x GetCandlesRequest_Duration) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GetCandlesRequest_Duration.Descriptor instead.
func (GetCandlesRequest_Duration) EnumDescriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{1, 0}
}

type Pair_Direction int32

const (
	Pair_UP   Pair_Direction = 0
	Pair_DOWN Pair_Direction = 1
)

// Enum value maps for Pair_Direction.
var (
	Pair_Direction_name = map[int32]string{
		0: "UP",
		1: "DOWN",
	}
	Pair_Direction_value = map[string]int32{
		"UP":   0,
		"DOWN": 1,
	}
)

func (x Pair_Direction) Enum() *Pair_Direction {
	p := new(Pair_Direction)
	*p = x
	return p
}

func (x Pair_Direction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Pair_Direction) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_moneytree_proto_enumTypes[1].Descriptor()
}

func (Pair_Direction) Type() protoreflect.EnumType {
	return &file_proto_moneytree_proto_enumTypes[1]
}

func (x Pair_Direction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Pair_Direction.Descriptor instead.
func (Pair_Direction) EnumDescriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{8, 0}
}

type NullRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NullRequest) Reset() {
	*x = NullRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NullRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NullRequest) ProtoMessage() {}

func (x *NullRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NullRequest.ProtoReflect.Descriptor instead.
func (*NullRequest) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{0}
}

type GetCandlesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Duration  GetCandlesRequest_Duration `protobuf:"varint,1,opt,name=duration,proto3,enum=moneytree.GetCandlesRequest_Duration" json:"duration,omitempty"`
	StartTime int64                      `protobuf:"varint,2,opt,name=startTime,proto3" json:"startTime,omitempty"`
	EndTime   int64                      `protobuf:"varint,3,opt,name=endTime,proto3" json:"endTime,omitempty"`
}

func (x *GetCandlesRequest) Reset() {
	*x = GetCandlesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCandlesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCandlesRequest) ProtoMessage() {}

func (x *GetCandlesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCandlesRequest.ProtoReflect.Descriptor instead.
func (*GetCandlesRequest) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{1}
}

func (x *GetCandlesRequest) GetDuration() GetCandlesRequest_Duration {
	if x != nil {
		return x.Duration
	}
	return GetCandlesRequest_ONE_MINUTE
}

func (x *GetCandlesRequest) GetStartTime() int64 {
	if x != nil {
		return x.StartTime
	}
	return 0
}

func (x *GetCandlesRequest) GetEndTime() int64 {
	if x != nil {
		return x.EndTime
	}
	return 0
}

type CandleCollection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Candles []*Candle `protobuf:"bytes,1,rep,name=candles,proto3" json:"candles,omitempty"`
}

func (x *CandleCollection) Reset() {
	*x = CandleCollection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CandleCollection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CandleCollection) ProtoMessage() {}

func (x *CandleCollection) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CandleCollection.ProtoReflect.Descriptor instead.
func (*CandleCollection) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{2}
}

func (x *CandleCollection) GetCandles() []*Candle {
	if x != nil {
		return x.Candles
	}
	return nil
}

type Candle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ts     int64  `protobuf:"varint,1,opt,name=ts,proto3" json:"ts,omitempty"`
	Open   string `protobuf:"bytes,2,opt,name=open,proto3" json:"open,omitempty"`
	Close  string `protobuf:"bytes,3,opt,name=close,proto3" json:"close,omitempty"`
	Low    string `protobuf:"bytes,4,opt,name=low,proto3" json:"low,omitempty"`
	High   string `protobuf:"bytes,5,opt,name=high,proto3" json:"high,omitempty"`
	Volume string `protobuf:"bytes,6,opt,name=volume,proto3" json:"volume,omitempty"`
}

func (x *Candle) Reset() {
	*x = Candle{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Candle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Candle) ProtoMessage() {}

func (x *Candle) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Candle.ProtoReflect.Descriptor instead.
func (*Candle) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{3}
}

func (x *Candle) GetTs() int64 {
	if x != nil {
		return x.Ts
	}
	return 0
}

func (x *Candle) GetOpen() string {
	if x != nil {
		return x.Open
	}
	return ""
}

func (x *Candle) GetClose() string {
	if x != nil {
		return x.Close
	}
	return ""
}

func (x *Candle) GetLow() string {
	if x != nil {
		return x.Low
	}
	return ""
}

func (x *Candle) GetHigh() string {
	if x != nil {
		return x.High
	}
	return ""
}

func (x *Candle) GetVolume() string {
	if x != nil {
		return x.Volume
	}
	return ""
}

type PlacePairRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Direction string `protobuf:"bytes,1,opt,name=direction,proto3" json:"direction,omitempty"`
}

func (x *PlacePairRequest) Reset() {
	*x = PlacePairRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PlacePairRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlacePairRequest) ProtoMessage() {}

func (x *PlacePairRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlacePairRequest.ProtoReflect.Descriptor instead.
func (*PlacePairRequest) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{4}
}

func (x *PlacePairRequest) GetDirection() string {
	if x != nil {
		return x.Direction
	}
	return ""
}

type PlacePairResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pair  *Pair  `protobuf:"bytes,1,opt,name=pair,proto3" json:"pair,omitempty"`
	Error *Error `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *PlacePairResponse) Reset() {
	*x = PlacePairResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PlacePairResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlacePairResponse) ProtoMessage() {}

func (x *PlacePairResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlacePairResponse.ProtoReflect.Descriptor instead.
func (*PlacePairResponse) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{5}
}

func (x *PlacePairResponse) GetPair() *Pair {
	if x != nil {
		return x.Pair
	}
	return nil
}

func (x *PlacePairResponse) GetError() *Error {
	if x != nil {
		return x.Error
	}
	return nil
}

type PairCollection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pairs []*Pair `protobuf:"bytes,1,rep,name=pairs,proto3" json:"pairs,omitempty"`
}

func (x *PairCollection) Reset() {
	*x = PairCollection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PairCollection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PairCollection) ProtoMessage() {}

func (x *PairCollection) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PairCollection.ProtoReflect.Descriptor instead.
func (*PairCollection) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{6}
}

func (x *PairCollection) GetPairs() []*Pair {
	if x != nil {
		return x.Pairs
	}
	return nil
}

type Order struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Side     string `protobuf:"bytes,1,opt,name=side,proto3" json:"side,omitempty"`
	Price    string `protobuf:"bytes,2,opt,name=price,proto3" json:"price,omitempty"`
	Quantity string `protobuf:"bytes,3,opt,name=quantity,proto3" json:"quantity,omitempty"`
	Filled   string `protobuf:"bytes,4,opt,name=filled,proto3" json:"filled,omitempty"`
	Status   string `protobuf:"bytes,5,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *Order) Reset() {
	*x = Order{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Order) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Order) ProtoMessage() {}

func (x *Order) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Order.ProtoReflect.Descriptor instead.
func (*Order) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{7}
}

func (x *Order) GetSide() string {
	if x != nil {
		return x.Side
	}
	return ""
}

func (x *Order) GetPrice() string {
	if x != nil {
		return x.Price
	}
	return ""
}

func (x *Order) GetQuantity() string {
	if x != nil {
		return x.Quantity
	}
	return ""
}

func (x *Order) GetFilled() string {
	if x != nil {
		return x.Filled
	}
	return ""
}

func (x *Order) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

type Pair struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uuid          string `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Created       int64  `protobuf:"varint,2,opt,name=created,proto3" json:"created,omitempty"`
	Ended         int64  `protobuf:"varint,3,opt,name=ended,proto3" json:"ended,omitempty"`
	Direction     string `protobuf:"bytes,4,opt,name=direction,proto3" json:"direction,omitempty"`
	Done          bool   `protobuf:"varint,5,opt,name=done,proto3" json:"done,omitempty"`
	Status        string `protobuf:"bytes,6,opt,name=status,proto3" json:"status,omitempty"`
	StatusDetails string `protobuf:"bytes,7,opt,name=statusDetails,proto3" json:"statusDetails,omitempty"`
	BuyOrder      *Order `protobuf:"bytes,8,opt,name=buyOrder,proto3" json:"buyOrder,omitempty"`
	SellOrder     *Order `protobuf:"bytes,9,opt,name=sellOrder,proto3" json:"sellOrder,omitempty"`
	ReversalOrder *Order `protobuf:"bytes,10,opt,name=reversalOrder,proto3" json:"reversalOrder,omitempty"`
}

func (x *Pair) Reset() {
	*x = Pair{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Pair) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Pair) ProtoMessage() {}

func (x *Pair) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Pair.ProtoReflect.Descriptor instead.
func (*Pair) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{8}
}

func (x *Pair) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *Pair) GetCreated() int64 {
	if x != nil {
		return x.Created
	}
	return 0
}

func (x *Pair) GetEnded() int64 {
	if x != nil {
		return x.Ended
	}
	return 0
}

func (x *Pair) GetDirection() string {
	if x != nil {
		return x.Direction
	}
	return ""
}

func (x *Pair) GetDone() bool {
	if x != nil {
		return x.Done
	}
	return false
}

func (x *Pair) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Pair) GetStatusDetails() string {
	if x != nil {
		return x.StatusDetails
	}
	return ""
}

func (x *Pair) GetBuyOrder() *Order {
	if x != nil {
		return x.BuyOrder
	}
	return nil
}

func (x *Pair) GetSellOrder() *Order {
	if x != nil {
		return x.SellOrder
	}
	return nil
}

func (x *Pair) GetReversalOrder() *Order {
	if x != nil {
		return x.ReversalOrder
	}
	return nil
}

type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_moneytree_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_proto_moneytree_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_proto_moneytree_proto_rawDescGZIP(), []int{9}
}

func (x *Error) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_proto_moneytree_proto protoreflect.FileDescriptor

var file_proto_moneytree_proto_rawDesc = []byte{
	0x0a, 0x15, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72,
	0x65, 0x65, 0x22, 0x0d, 0x0a, 0x0b, 0x4e, 0x75, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x22, 0x88, 0x02, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x43, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x41, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x25, 0x2e, 0x6d, 0x6f, 0x6e, 0x65,
	0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x64, 0x54,
	0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x65, 0x6e, 0x64, 0x54, 0x69,
	0x6d, 0x65, 0x22, 0x78, 0x0a, 0x08, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0e,
	0x0a, 0x0a, 0x4f, 0x4e, 0x45, 0x5f, 0x4d, 0x49, 0x4e, 0x55, 0x54, 0x45, 0x10, 0x00, 0x12, 0x10,
	0x0a, 0x0c, 0x46, 0x49, 0x56, 0x45, 0x5f, 0x4d, 0x49, 0x4e, 0x55, 0x54, 0x45, 0x53, 0x10, 0x01,
	0x12, 0x13, 0x0a, 0x0f, 0x46, 0x49, 0x46, 0x54, 0x45, 0x45, 0x4e, 0x5f, 0x4d, 0x49, 0x4e, 0x55,
	0x54, 0x45, 0x53, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x4f, 0x4e, 0x45, 0x5f, 0x48, 0x4f, 0x55,
	0x52, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c, 0x54, 0x57, 0x45, 0x4c, 0x56, 0x45, 0x5f, 0x48, 0x4f,
	0x55, 0x52, 0x53, 0x10, 0x04, 0x12, 0x15, 0x0a, 0x11, 0x54, 0x57, 0x45, 0x4e, 0x54, 0x59, 0x5f,
	0x46, 0x4f, 0x55, 0x52, 0x5f, 0x48, 0x4f, 0x55, 0x52, 0x53, 0x10, 0x05, 0x22, 0x3f, 0x0a, 0x10,
	0x43, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x2b, 0x0a, 0x07, 0x63, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x11, 0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x43, 0x61,
	0x6e, 0x64, 0x6c, 0x65, 0x52, 0x07, 0x63, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73, 0x22, 0x80, 0x01,
	0x0a, 0x06, 0x43, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x74, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6f, 0x70, 0x65, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6f, 0x70, 0x65, 0x6e, 0x12, 0x14, 0x0a, 0x05,
	0x63, 0x6c, 0x6f, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x63, 0x6c, 0x6f,
	0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x77, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6c, 0x6f, 0x77, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x69, 0x67, 0x68, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x68, 0x69, 0x67, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x6f, 0x6c, 0x75,
	0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65,
	0x22, 0x30, 0x0a, 0x10, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x50, 0x61, 0x69, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x22, 0x60, 0x0a, 0x11, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x50, 0x61, 0x69, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x23, 0x0a, 0x04, 0x70, 0x61, 0x69, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65,
	0x65, 0x2e, 0x50, 0x61, 0x69, 0x72, 0x52, 0x04, 0x70, 0x61, 0x69, 0x72, 0x12, 0x26, 0x0a, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x6f,
	0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x22, 0x37, 0x0a, 0x0e, 0x50, 0x61, 0x69, 0x72, 0x43, 0x6f, 0x6c, 0x6c,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x05, 0x70, 0x61, 0x69, 0x72, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65,
	0x65, 0x2e, 0x50, 0x61, 0x69, 0x72, 0x52, 0x05, 0x70, 0x61, 0x69, 0x72, 0x73, 0x22, 0x7d, 0x0a,
	0x05, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x64, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x73, 0x69, 0x64, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72,
	0x69, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x06,
	0x66, 0x69, 0x6c, 0x6c, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x66, 0x69,
	0x6c, 0x6c, 0x65, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0xef, 0x02, 0x0a,
	0x04, 0x50, 0x61, 0x69, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6e, 0x64, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x05, 0x65, 0x6e, 0x64, 0x65, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x69, 0x72,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x64, 0x69,
	0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x44, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x2c, 0x0a, 0x08, 0x62, 0x75, 0x79,
	0x4f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x6f,
	0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x52, 0x08, 0x62,
	0x75, 0x79, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x2e, 0x0a, 0x09, 0x73, 0x65, 0x6c, 0x6c, 0x4f,
	0x72, 0x64, 0x65, 0x72, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x6f, 0x6e,
	0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x52, 0x09, 0x73, 0x65,
	0x6c, 0x6c, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x36, 0x0a, 0x0d, 0x72, 0x65, 0x76, 0x65, 0x72,
	0x73, 0x61, 0x6c, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10,
	0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x4f, 0x72, 0x64, 0x65, 0x72,
	0x52, 0x0d, 0x72, 0x65, 0x76, 0x65, 0x72, 0x73, 0x61, 0x6c, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x22,
	0x1d, 0x0a, 0x09, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x06, 0x0a, 0x02,
	0x55, 0x50, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x01, 0x22, 0x21,
	0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x32, 0xdf, 0x01, 0x0a, 0x09, 0x4d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x12,
	0x46, 0x0a, 0x09, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x50, 0x61, 0x69, 0x72, 0x12, 0x1b, 0x2e, 0x6d,
	0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x50, 0x61,
	0x69, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x6d, 0x6f, 0x6e, 0x65,
	0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x50, 0x61, 0x69, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x41, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x4f, 0x70,
	0x65, 0x6e, 0x50, 0x61, 0x69, 0x72, 0x73, 0x12, 0x16, 0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74,
	0x72, 0x65, 0x65, 0x2e, 0x4e, 0x75, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x19, 0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x50, 0x61, 0x69, 0x72,
	0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x47, 0x0a, 0x0a, 0x47, 0x65,
	0x74, 0x43, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73, 0x12, 0x1c, 0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79,
	0x74, 0x72, 0x65, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72,
	0x65, 0x65, 0x2e, 0x43, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x42, 0x5f, 0x0a, 0x1c, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x69, 0x6e, 0x69, 0x6d,
	0x69, 0x6e, 0x69, 0x2e, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x42, 0x0e, 0x4d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x73, 0x69, 0x6e, 0x69, 0x73, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x2f, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x74, 0x72, 0x65, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_moneytree_proto_rawDescOnce sync.Once
	file_proto_moneytree_proto_rawDescData = file_proto_moneytree_proto_rawDesc
)

func file_proto_moneytree_proto_rawDescGZIP() []byte {
	file_proto_moneytree_proto_rawDescOnce.Do(func() {
		file_proto_moneytree_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_moneytree_proto_rawDescData)
	})
	return file_proto_moneytree_proto_rawDescData
}

var file_proto_moneytree_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_proto_moneytree_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_proto_moneytree_proto_goTypes = []interface{}{
	(GetCandlesRequest_Duration)(0), // 0: moneytree.GetCandlesRequest.Duration
	(Pair_Direction)(0),             // 1: moneytree.Pair.Direction
	(*NullRequest)(nil),             // 2: moneytree.NullRequest
	(*GetCandlesRequest)(nil),       // 3: moneytree.GetCandlesRequest
	(*CandleCollection)(nil),        // 4: moneytree.CandleCollection
	(*Candle)(nil),                  // 5: moneytree.Candle
	(*PlacePairRequest)(nil),        // 6: moneytree.PlacePairRequest
	(*PlacePairResponse)(nil),       // 7: moneytree.PlacePairResponse
	(*PairCollection)(nil),          // 8: moneytree.PairCollection
	(*Order)(nil),                   // 9: moneytree.Order
	(*Pair)(nil),                    // 10: moneytree.Pair
	(*Error)(nil),                   // 11: moneytree.Error
}
var file_proto_moneytree_proto_depIdxs = []int32{
	0,  // 0: moneytree.GetCandlesRequest.duration:type_name -> moneytree.GetCandlesRequest.Duration
	5,  // 1: moneytree.CandleCollection.candles:type_name -> moneytree.Candle
	10, // 2: moneytree.PlacePairResponse.pair:type_name -> moneytree.Pair
	11, // 3: moneytree.PlacePairResponse.error:type_name -> moneytree.Error
	10, // 4: moneytree.PairCollection.pairs:type_name -> moneytree.Pair
	9,  // 5: moneytree.Pair.buyOrder:type_name -> moneytree.Order
	9,  // 6: moneytree.Pair.sellOrder:type_name -> moneytree.Order
	9,  // 7: moneytree.Pair.reversalOrder:type_name -> moneytree.Order
	6,  // 8: moneytree.Moneytree.PlacePair:input_type -> moneytree.PlacePairRequest
	2,  // 9: moneytree.Moneytree.GetOpenPairs:input_type -> moneytree.NullRequest
	3,  // 10: moneytree.Moneytree.GetCandles:input_type -> moneytree.GetCandlesRequest
	7,  // 11: moneytree.Moneytree.PlacePair:output_type -> moneytree.PlacePairResponse
	8,  // 12: moneytree.Moneytree.GetOpenPairs:output_type -> moneytree.PairCollection
	4,  // 13: moneytree.Moneytree.GetCandles:output_type -> moneytree.CandleCollection
	11, // [11:14] is the sub-list for method output_type
	8,  // [8:11] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_proto_moneytree_proto_init() }
func file_proto_moneytree_proto_init() {
	if File_proto_moneytree_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_moneytree_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NullRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCandlesRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CandleCollection); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Candle); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PlacePairRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PlacePairResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PairCollection); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Order); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Pair); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_moneytree_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_moneytree_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_moneytree_proto_goTypes,
		DependencyIndexes: file_proto_moneytree_proto_depIdxs,
		EnumInfos:         file_proto_moneytree_proto_enumTypes,
		MessageInfos:      file_proto_moneytree_proto_msgTypes,
	}.Build()
	File_proto_moneytree_proto = out.File
	file_proto_moneytree_proto_rawDesc = nil
	file_proto_moneytree_proto_goTypes = nil
	file_proto_moneytree_proto_depIdxs = nil
}
