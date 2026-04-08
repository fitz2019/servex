package cqrs

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateOrderCommand 测试用命令.
type CreateOrderCommand struct {
	UserID    string
	ProductID string
	Quantity  int
}

// CreateOrderResult 测试用命令结果.
type CreateOrderResult struct {
	OrderID string
}

// createOrderHandler 测试用命令处理器.
type createOrderHandler struct {
	shouldFail bool
}

func (h *createOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) (CreateOrderCommand, CreateOrderResult, error) {
	if h.shouldFail {
		return cmd, CreateOrderResult{}, errors.New("create order failed")
	}
	return cmd, CreateOrderResult{OrderID: "ORD-" + cmd.UserID + "-" + cmd.ProductID}, nil
}

func TestApplyCommand_Success(t *testing.T) {
	ctx := t.Context()
	handler := &createOrderHandler{}

	cmd := CreateOrderCommand{
		UserID:    "user-1",
		ProductID: "prod-1",
		Quantity:  5,
	}

	returnedCmd, result, err := ApplyCommand(ctx, cmd, handler)

	require.NoError(t, err)
	assert.Equal(t, cmd, returnedCmd)
	assert.Equal(t, "ORD-user-1-prod-1", result.OrderID)
}

func TestApplyCommand_Error(t *testing.T) {
	ctx := t.Context()
	handler := &createOrderHandler{shouldFail: true}

	cmd := CreateOrderCommand{
		UserID:    "user-1",
		ProductID: "prod-1",
		Quantity:  5,
	}

	returnedCmd, result, err := ApplyCommand(ctx, cmd, handler)

	require.Error(t, err)
	assert.Equal(t, "create order failed", err.Error())
	assert.Equal(t, cmd, returnedCmd)
	assert.Empty(t, result.OrderID)
}

func TestApplyCommand_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	handler := &contextAwareCommandHandler{}

	cmd := CreateOrderCommand{UserID: "user-1"}

	_, _, err := ApplyCommand(ctx, cmd, handler)

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestChainCommand(t *testing.T) {
	var order []string

	mw1 := func(next CommandHandler[CreateOrderCommand, CreateOrderResult]) CommandHandler[CreateOrderCommand, CreateOrderResult] {
		return &commandHandlerFunc[CreateOrderCommand, CreateOrderResult]{
			fn: func(ctx context.Context, cmd CreateOrderCommand) (CreateOrderCommand, CreateOrderResult, error) {
				order = append(order, "mw1:before")
				cmd, result, err := next.Handle(ctx, cmd)
				order = append(order, "mw1:after")
				return cmd, result, err
			},
		}
	}

	mw2 := func(next CommandHandler[CreateOrderCommand, CreateOrderResult]) CommandHandler[CreateOrderCommand, CreateOrderResult] {
		return &commandHandlerFunc[CreateOrderCommand, CreateOrderResult]{
			fn: func(ctx context.Context, cmd CreateOrderCommand) (CreateOrderCommand, CreateOrderResult, error) {
				order = append(order, "mw2:before")
				cmd, result, err := next.Handle(ctx, cmd)
				order = append(order, "mw2:after")
				return cmd, result, err
			},
		}
	}

	handler := &createOrderHandler{}
	chained := ChainCommand[CreateOrderCommand, CreateOrderResult](handler, mw1, mw2)

	cmd := CreateOrderCommand{UserID: "u1", ProductID: "p1", Quantity: 1}
	_, result, err := chained.Handle(t.Context(), cmd)

	require.NoError(t, err)
	assert.Equal(t, "ORD-u1-p1", result.OrderID)
	assert.Equal(t, []string{"mw1:before", "mw2:before", "mw2:after", "mw1:after"}, order)
}

func TestChainCommand_ErrorPropagation(t *testing.T) {
	mw := func(next CommandHandler[CreateOrderCommand, CreateOrderResult]) CommandHandler[CreateOrderCommand, CreateOrderResult] {
		return &commandHandlerFunc[CreateOrderCommand, CreateOrderResult]{
			fn: func(ctx context.Context, cmd CreateOrderCommand) (CreateOrderCommand, CreateOrderResult, error) {
				return next.Handle(ctx, cmd)
			},
		}
	}

	handler := &createOrderHandler{shouldFail: true}
	chained := ChainCommand[CreateOrderCommand, CreateOrderResult](handler, mw)

	_, _, err := chained.Handle(t.Context(), CreateOrderCommand{UserID: "u1"})
	require.Error(t, err)
	assert.Equal(t, "create order failed", err.Error())
}

// contextAwareCommandHandler 检查 context 的处理器.
type contextAwareCommandHandler struct{}

func (h *contextAwareCommandHandler) Handle(ctx context.Context, cmd CreateOrderCommand) (CreateOrderCommand, CreateOrderResult, error) {
	if err := ctx.Err(); err != nil {
		return cmd, CreateOrderResult{}, err
	}
	return cmd, CreateOrderResult{OrderID: "ORD-123"}, nil
}
