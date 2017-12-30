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

func (p *Person) PublicKey() (p2 *Person) {
	p2 = new(Person)
	p2.Keys, _ = keypair.NewFromPublic(p.Keys.Public)
	return
}

func (p *Person) Public() string {
	return p.Keys.Public
}

func FromPublicKey(publicKey string) (p *Person, err error) {
	p = new(Person)
	p.Keys, err = keypair.NewFromPublic(publicKey)
	return
}

func FromPublicPrivateKeys(publicKey, privateKey string) (p *Person, err error) {
	p = new(Person)
	p.Keys, err = keypair.NewFromPair(publicKey, privateKey)
	return
}
