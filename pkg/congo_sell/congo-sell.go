package congo_sell

import (
	"embed"
	"log"

	"github.com/ccutch/congo/pkg/congo"
)

//go:embed all:migrations
var migrations embed.FS

type CongoSell struct {
	db      *congo.Database
	backend Backend
}

func InitCongoSell(root string, opts ...CongoSellOpts) *CongoSell {
	db := congo.SetupDatabase(root, "sell.db", migrations)
	if err := db.MigrateUp(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	db.Query(` UPDATE products SET active = false `)

	c := CongoSell{db, nil}
	for _, o := range opts {
		o(&c)
	}

	return &c
}

type CongoSellOpts func(*CongoSell)

func WithBackend(backend Backend) CongoSellOpts {
	return func(c *CongoSell) {
		c.backend = backend
	}
}

func WithProduct(name, description string, price int) CongoSellOpts {
	return func(c *CongoSell) {
		var p Product
		pi, err := c.Product(name)
		if err != nil {
			if p, err = c.backend.CreateProduct(name, description); err != nil {
				log.Fatalf("Failed to create product: %s", err)
			}
			if _, err = p.SetPrice(price); err != nil {
				log.Fatalf("Failed to set price for %s: %s", name, err)
			}
		} else {
			p, err = pi.Product()
		}

		if err != nil {
			log.Fatal("Failed to create product:", err)
		}

		err = c.db.Query(`

			INSERT INTO products (id, name, description, price, active)
			VALUES (?, ?, ?, ?, true)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				price = EXCLUDED.price,
				active = EXCLUDED.active,
				updated_at = CURRENT_TIMESTAMP

		`, p.ID(), name, description, price).Exec()
		if err != nil {
			log.Println("Failed to create product:", err)
		}
	}
}

func (c *CongoSell) Products() (products []*ProductInfo, err error) {
	return products, c.db.Query(`

		SELECT id, name, description, price, created_at, updated_at
		FROM products
		WHERE active = true

	`).All(func(scan congo.Scanner) error {
		p := ProductInfo{CongoSell: c, Model: c.db.Model()}
		products = append(products, &p)
		return scan(&p.ID, &p.Name, &p.Description, &p.PriceAmount, &p.CreatedAt, &p.UpdatedAt)
	})
}

func (c *CongoSell) Product(ident string) (*ProductInfo, error) {
	pi := ProductInfo{CongoSell: c, Model: c.db.Model()}
	return &pi, c.db.Query(`
	
		SELECT id, name, description, price, created_at, updated_at
		FROM products
		WHERE id = $1 OR name = $1
	
	`, ident).Scan(&pi.ID, &pi.Name, &pi.Description, &pi.PriceAmount, &pi.CreatedAt, &pi.UpdatedAt)
}

type ProductInfo struct {
	*CongoSell
	congo.Model
	Name        string
	Description string
	PriceAmount int
}

func (pi *ProductInfo) Product() (Product, error) {
	return pi.CongoSell.backend.GetProduct(pi.ID)
}

func (pi *ProductInfo) Price() (Price, error) {
	p, err := pi.Product()
	if err != nil {
		return nil, err
	}
	return p.Price()
}

func (pi *ProductInfo) Checkout(redirect string) (string, error) {
	p, err := pi.Price()
	if err != nil {
		return "", err
	}
	return p.CheckoutURL(redirect)
}
