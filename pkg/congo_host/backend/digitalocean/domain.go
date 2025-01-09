package digitalocean

type Domain struct {
	server     *Server
	DomainName string
}

func (d *Domain) Verify(other ...string) error {
	panic("unimplemented")
}
