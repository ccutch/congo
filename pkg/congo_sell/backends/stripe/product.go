package stripe

import (
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

func (p *Product) SetPrice(amount int) (err error) {
	prices := price.List(&stripe.PriceListParams{
		Product: stripe.String(p.ID()),
		Currency: stripe.String("usd"),
	})
	if prices.Next() {
		prices.Price()
		_, err = price.Update(prices.Price().ID, &stripe.PriceParams{
			Product:    stripe.String(p.ID()),
			UnitAmount: stripe.Int64(int64(amount)),
			Recurring: &stripe.PriceRecurringParams{
				Interval: stripe.String("month"),
			},
		})
		return err
	} else {
		_, err = price.New(&stripe.PriceParams{
			Product:    stripe.String(p.ID()),
			Currency:   stripe.String("usd"),
			UnitAmount: stripe.Int64(int64(amount)),
			Recurring: &stripe.PriceRecurringParams{
				Interval: stripe.String("month"),
			},
		})
	}
	return err
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
