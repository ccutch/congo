package stripe

import (
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentlink"
)

type Price struct {
	client *Client
	*stripe.Price
}

func (p *Price) ID() string {
	return p.Price.ID
}

func (p *Price) Amount() int {
	return int(p.Price.UnitAmount)
}

func (p *Price) Currency() string {
	return string(p.Price.Currency)
}

func (p *Price) CheckoutURL(redirect string) (string, error) {
	params := &stripe.PaymentLinkParams{
		LineItems: []*stripe.PaymentLinkLineItemParams{
			{
				Price:    stripe.String(p.ID()),
				Quantity: stripe.Int64(1),
			},
		},
		AfterCompletion: &stripe.PaymentLinkAfterCompletionParams{
			Type: stripe.String("redirect"),
			Redirect: &stripe.PaymentLinkAfterCompletionRedirectParams{
				URL: stripe.String(redirect),
			},
		},
	}
	result, err := paymentlink.New(params)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}
