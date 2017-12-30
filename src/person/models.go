package person

import "github.com/schollz/kiki/src/keypair"

// Person is just a set of keys
type Person struct {
	Keys *keypair.KeyPair `json:"keys"`
}

func New() (p *Person, err error) {
	p = new(Person)
	p.Keys, err = keypair.New()
	return
}
