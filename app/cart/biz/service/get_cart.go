package service

import (
	"context"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/qitian118/gomall/app/cart/biz/dal/mysql"
	"github.com/qitian118/gomall/app/cart/biz/model"
	cart "github.com/qitian118/gomall/rpc_gen/kitex_gen/cart"
)

type GetCartService struct {
	ctx context.Context
} // NewGetCartService new GetCartService
func NewGetCartService(ctx context.Context) *GetCartService {
	return &GetCartService{ctx: ctx}
}

// Run create note info
func (s *GetCartService) Run(req *cart.GetCartReq) (resp *cart.GetCartResp, err error) {
	// Finish your business logic.
	// fmt.Println("before GetCartByUserId")
	list, err := model.GetCartByUserId(s.ctx, mysql.DB, req.UserId)
	if err != nil {
		return nil, kerrors.NewBizStatusError(50002, err.Error())
	}
	var items []*cart.CartItem
	// fmt.Println("before range")
	for _, item := range list {
		items = append(items, &cart.CartItem{
			ProductId: item.ProductId,
			Quantity:  item.Qty,
		})
	}
	// fmt.Println("before return")
	// fmt.Println(items)
	return &cart.GetCartResp{Items: items}, nil
}
