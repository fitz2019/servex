package errors

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcDetail gRPC Status detail 中携带的错误详情.
type grpcDetail struct {
	Code     int               `json:"code"`
	Key      string            `json:"key"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitzero"`
}

// ToGRPCStatus 将 error 转为 gRPC Status.
func ToGRPCStatus(err error) *status.Status {
	if err == nil {
		return status.New(codes.OK, "")
	}

	e, ok := FromError(err)
	if !ok {
		return status.New(codes.Internal, err.Error())
	}

	code := e.GRPC
	if code == codes.OK {
		code = codes.Internal
	}

	detail := &grpcDetail{
		Code:     e.Code,
		Key:      e.Key,
		Message:  e.Message,
		Metadata: e.Metadata,
	}
	detailJSON, _ := json.Marshal(detail)

	return status.New(code, string(detailJSON))
}

// FromGRPCStatus 从 gRPC Status 还原 *Error.
func FromGRPCStatus(st *status.Status) *Error {
	if st == nil || st.Code() == codes.OK {
		return nil
	}

	var detail grpcDetail
	if err := json.Unmarshal([]byte(st.Message()), &detail); err == nil && detail.Code != 0 {
		return &Error{
			Code:     detail.Code,
			Key:      detail.Key,
			Message:  detail.Message,
			Metadata: detail.Metadata,
			GRPC:     st.Code(),
		}
	}

	return &Error{
		Code:    int(st.Code()),
		Key:     st.Code().String(),
		Message: st.Message(),
		GRPC:    st.Code(),
	}
}

// UnaryServerInterceptor 返回 gRPC 一元拦截器.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		return nil, ToGRPCStatus(err).Err()
	}
}

// StreamServerInterceptor 返回 gRPC 流式拦截器.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := handler(srv, ss)
		if err == nil {
			return nil
		}
		return ToGRPCStatus(err).Err()
	}
}
