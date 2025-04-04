package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Koch13o1/go-grpc-graphql-microservice/order"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
)

type mutationResolver struct {
	server *Server
}

// CreateAccount
func (r *mutationResolver) CreateAccount(ctx context.Context, in AccountInput) (*Account, error) {
	_, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	a, err := r.server.accountClient.PostAccount(ctx, in.Name)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Account{
		ID:   a.ID,
		Name: a.Name,
	}, nil
}

// CreateProduct
func (r *mutationResolver) CreateProduct(ctx context.Context, in ProductInput) (*Product, error) {
	_, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	p, err := r.server.catalogClient.PostProduct(ctx, in.Name, in.Description, in.Price)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, nil
}

// CreateOrder
func (r *mutationResolver) CreateOrder(ctx context.Context, in OrderInput) (*Order, error) {
	_, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var products []order.OrderedProduct
	for _, p := range in.Products {
		if p.Quantity <= 0 {
			return nil, ErrInvalidParameter
		}
		products = append(products, order.OrderedProduct{
			ID:       p.ID,
			Quantity: uint32(p.Quantity),
		})
	}
	o, err := r.server.orderClient.PostOrder(ctx, in.AccountID, products)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &Order{
		ID:         o.ID,
		CreatedAt:  o.CreatedAt,
		TotalPrice: o.TotalPrice,
	}, nil
}
