package congo_sell

type Backend interface {
	CreateProduct(name, description string) (Product, error)
	GetProduct(id string) (Product, error)
	GetProductByName(name string) (Product, error)
}

type Product interface {
	ID() string
	Name() string
	Description() string
	Update(name, description string) error
	Delete() error
	Price() (Price, error)
	SetPrice(int) (Price, error)
}

type Price interface {
	ID() string
	Amount() int
	Currency() string
	CheckoutURL(string) (string, error)
}
