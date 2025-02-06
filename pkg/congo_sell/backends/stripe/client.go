package stripe

import (
	"errors"

	"github.com/ccutch/congo/pkg/congo_sell"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/product"
)

type Client struct {
	key string
}

func NewClient(key string) *Client {
	if key == "" {
		return nil
	}
	stripe.Key = key
	return &Client{key}
}

func (c *Client) CreateProduct(name, description string) (congo_sell.Product, error) {
	result, err := product.New(&stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		Active:      stripe.Bool(true),
		Type:        stripe.String("service"),
	})
	if err != nil {
		return nil, err
	}
	return &Product{c, result}, nil
}

// GetProducts returns all products (not required by congo_sell.Backend)
func (c *Client) GetProducts() (ret []congo_sell.Product, err error) {
	cursor := product.List(&stripe.ProductListParams{})
	for cursor.Next() {
		product := cursor.Product()
		ret = append(ret, &Product{c, product})
	}
	return ret, cursor.Err()
}

func (c *Client) GetProduct(id string) (congo_sell.Product, error) {
	product, err := product.Get(id, nil)
	if err != nil {
		return nil, err
	}
	return &Product{c, product}, nil
}

func (c *Client) GetProductByName(name string) (congo_sell.Product, error) {
	products, err := c.GetProducts()
	if err != nil {
		return nil, err
	}
	for _, p := range products {
		return p, nil
	}
	return nil, errors.New("product not found")
}
