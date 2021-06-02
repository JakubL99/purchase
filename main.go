package main

import (
	cartProto "purchase/cartProto"
	handler "purchase/handler"
	ordersProto "purchase/ordersProto"
	pb "purchase/proto"
	supplyProto "purchase/supplyProto"
	usersProto "purchase/usersProto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
)

func main() {

	srv := service.New(
		service.Name("purchase"),
		service.Version("latest"),
	)

	srv.Init()

	cartClient := cartProto.NewCartService("cart", srv.Client())
	supplyClient := supplyProto.NewSupplyService("supply", srv.Client())
	usersClient := usersProto.NewUsersService("users", srv.Client())
	ordersClient := ordersProto.NewOrdersService("orders", srv.Client())

	h := &handler.Handler{
		CartClient:   cartClient,
		SupplyClient: supplyClient,
		UsersClient:  usersClient,
		OrdersClient: ordersClient,
	}

	pb.RegisterPurchaseHandler(srv.Server(), h)

	if err := srv.Run(); err != nil {
		logger.Error(err)
	}
}
