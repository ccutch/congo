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
	SetPrice(int) error
	Update(name, description string) error
	Delete() error
}
