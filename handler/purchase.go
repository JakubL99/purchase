package handler

import (
	"context"
	cartProto "purchase/cartProto"
	ordersProto "purchase/ordersProto"
	pb "purchase/proto"
	supplyProto "purchase/supplyProto"
	usersProto "purchase/usersProto"
	"strconv"

	"github.com/micro/micro/v3/service/logger"
)

type Repository interface {
}

type Handler struct {
	Repository
	CartClient   cartProto.CartService
	SupplyClient supplyProto.SupplyService
	UsersClient  usersProto.UsersService
	OrdersClient ordersProto.OrdersService
}

func MarshalGetSave(id *pb.User) *cartProto.ShoppingCart {
	return &cartProto.ShoppingCart{
		IdUser: id.IdUser,
	}
}

func MarshalGetUser(id *pb.User) *usersProto.User {
	return &usersProto.User{
		Id: id.IdUser,
	}
}

func MarshalSupplyPrice(price string) *supplyProto.Price {
	return &supplyProto.Price{
		Price: price,
	}
}

func MarshalProducts(products []*cartProto.Product) []*pb.Product {
	collection := make([]*pb.Product, 0)
	for _, product := range products {
		collection = append(collection, MarshalProduct(product))
	}
	return collection
}

func MarshalProduct(product *cartProto.Product) *pb.Product {
	return &pb.Product{
		IdProduct: product.IdProduct,
		Name:      product.Name,
		Price:     product.Price,
	}
}

func MarshalAddress(req *pb.Details) *ordersProto.Address {
	return &ordersProto.Address{
		Country: req.Address.Country,
		City:    req.Address.City,
		Post:    req.Address.Post,
		Street:  req.Address.Street,
		Number:  req.Address.Number,
	}
}

func MarshalOrderProduct(product *cartProto.Product) *ordersProto.Product {
	return &ordersProto.Product{
		IdProduct: product.IdProduct,
		Name:      product.Name,
		Price:     product.Price,
	}
}

func MarshalCollectionProducts(products []*cartProto.Product) []*ordersProto.Product {
	collection := make([]*ordersProto.Product, 0)
	for _, product := range products {
		collection = append(collection, MarshalOrderProduct(product))
	}
	return collection
}

func convertIdUser(id *pb.Details) *pb.User {
	return &pb.User{
		IdUser: id.IdUser,
	}
}

func (h *Handler) Start(ctx context.Context, req *pb.User, rsp *pb.Purch) error {
	cartResponse, err := h.CartClient.GetCart(ctx, MarshalGetSave(req))
	if err != nil {
		logger.Error("Error Download Cart: ", err)
	}
	logger.Info("cartResponse: ", cartResponse)

	var products []*cartProto.Product
	products = cartResponse.Products
	logger.Info("Products: ", products)

	var price float64
	price = 0.00

	for _, v := range products {
		p := v.Price
		floatPrice, err := strconv.ParseFloat(p, 64)
		if err != nil {
			logger.Error("Error strconv string to Float:  ", err)
		}
		price = price + floatPrice
	}

	logger.Info("price: ", price)
	priceString := strconv.FormatFloat(price, 'f', 2, 64)
	supplyResponse, err := h.SupplyClient.Calculate(ctx, MarshalSupplyPrice(priceString))
	rsp.PriceSupply = supplyResponse.Expense
	rsp.PriceOrder = priceString
	rsp.Products = MarshalProducts(cartResponse.Products)
	return nil
}

func (h *Handler) Implementation(ctx context.Context, req *pb.Details, rsp *pb.Order) error {
	id := convertIdUser(req)
	cartResponse, err := h.CartClient.GetCart(ctx, MarshalGetSave(id))
	if err != nil {
		logger.Error("Error Download Cart: ", err)
	}
	logger.Info("cartResponse: ", cartResponse)

	var products []*cartProto.Product
	products = cartResponse.Products
	logger.Info("Products: ", products)

	var price float64
	price = 0.00

	for _, v := range products {
		p := v.Price
		floatPrice, err := strconv.ParseFloat(p, 64)
		if err != nil {
			logger.Error("Error strconv string to Float:  ", err)
		}
		price = price + floatPrice
	}
	usersResponse, err := h.UsersClient.Get(ctx, MarshalGetUser(id))
	logger.Info("usersResponse", usersResponse)

	priceString := strconv.FormatFloat(price, 'f', 2, 64)
	supplyResponse, err := h.SupplyClient.Calculate(ctx, MarshalSupplyPrice(priceString))

	sr, err := strconv.ParseFloat(supplyResponse.Expense, 64)
	if err != nil {
		logger.Error("Error strconv string to Float:  ", err)
	}
	priceOrder := sr + price
	logger.Info("priceOrder ", priceOrder)

	var order ordersProto.Order
	var address *ordersProto.Address

	address = MarshalAddress(req)
	order.Products = MarshalCollectionProducts(products)
	order.Price = strconv.FormatFloat(priceOrder, 'f', 2, 64)
	order.IdUser = id.IdUser
	order.Name = usersResponse.User.Name
	order.Surname = usersResponse.User.Surname
	order.Address = address
	order.Status = "Is not paid"

	logger.Info("Order: ", order)
	var o *ordersProto.Order
	o = &order
	ordersResponse, err := h.OrdersClient.Create(ctx, o)
	if err != nil {
		logger.Error("Error Create Order: ", err)
	}
	logger.Info("ordersResponse: ", ordersResponse)

	rsp.Name = order.Name
	rsp.Surname = order.Surname
	rsp.NumberAccount = "45 5555 5555 5555 1234 1234 "
	rsp.Price = order.Price
	rsp.NumberOrder = ordersResponse.NumberOrder

	cartResponseDelete, err := h.CartClient.DeleteCart(ctx, MarshalGetSave(id))
	if err != nil {
		logger.Error("Error delete cart", err)
	}
	logger.Info("Delete cart: ", cartResponseDelete)
	return nil
}
