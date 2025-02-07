package stripe

import (
	"errors"

	"github.com/ccutch/congo/pkg/congo_sell"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
)

type Product struct {
	client *Client
	*stripe.Product
}

func (p *Product) ID() string {
	return p.Product.ID
}

func (p *Product) Name() string {
	return p.Product.Name
}

func (p *Product) Description() string {
	return p.Product.Description
}

func (p *Product) Update(name, description string) (err error) {
	stripe.Key = p.client.key
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
	}
	p.Product, err = product.Update(p.ID(), params)
	return err
}

func (p *Product) Delete() error {
	stripe.Key = p.client.key
	_, err := product.Del(p.ID(), nil)
	return err
}

func (p *Product) Price() (congo_sell.Price, error) {
	prices := price.List(&stripe.PriceListParams{
		Product: stripe.String(p.ID()),
		Active:  stripe.Bool(true),
	})
	if prices.Next() {
		return &Price{p.client, prices.Price()}, nil
	}
	return nil, errors.New("no price found")
}

func (p *Product) SetPrice(amount int) (congo_sell.Price, error) {
	if old, err := p.Price(); err == nil {
		_, err := price.Update(old.ID(), &stripe.PriceParams{
			Active: stripe.Bool(false),
		})
		if err != nil {
			return nil, err
		}
	}
	pr, err := price.New(&stripe.PriceParams{
		Product:    stripe.String(p.ID()),
		Currency:   stripe.String("usd"),
		UnitAmount: stripe.Int64(int64(amount)),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String("month"),
		},
	})
	return &Price{p.client, pr}, err
}
