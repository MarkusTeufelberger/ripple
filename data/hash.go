package data

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/donovanhide/ripple/crypto"
)

type Hash128 [16]byte
type Hash160 [20]byte
type Hash256 [32]byte
type Vector256 []Hash256
type VariableLength []byte
type PublicKey [33]byte
type Account [20]byte
type RegularKey [20]byte

var zero256 Hash256
var zeroAccount Account
var zeroPublicKey PublicKey

func (h Hash128) Bytes() []byte {
	return h[:]
}

func (h Hash128) String() string {
	return string(b2h(h[:]))
}

func (h Hash160) Bytes() []byte {
	return h[:]
}

func (h Hash160) String() string {
	return string(b2h(h[:]))
}

// Accepts either a hex string or a byte slice of length 32
func NewHash256(value interface{}) (*Hash256, error) {
	var h Hash256
	switch v := value.(type) {
	case []byte:
		if len(v) != 32 {
			return nil, fmt.Errorf("NewHash256: Wrong length %X", value)
		}
		copy(h[:], v)
	case string:
		n, err := hex.Decode(h[:], []byte(v))
		if err != nil {
			return nil, err
		}
		if n != 32 {
			return nil, fmt.Errorf("NewHash256: Wrong length %s", v)
		}
	default:
		return nil, fmt.Errorf("NewHash256: Wrong type %+v", v)
	}
	return &h, nil
}

func (h Hash256) IsZero() bool {
	return h == zero256
}

func (h Hash256) Xor(x Hash256) Hash256 {
	var xor Hash256
	for i := range h {
		xor[i] = h[i] ^ x[i]
	}
	return x
}

func (h Hash256) Compare(x Hash256) int {
	return bytes.Compare(h[:], x[:])
}

func (h Hash256) Bytes() []byte {
	return h[:]
}

func (h Hash256) String() string {
	return string(b2h(h[:]))
}

func (h Hash256) TruncatedString(length int) string {
	return string(b2h(h[:length]))
}

func (v *VariableLength) String() string {
	if v != nil {
		b, _ := v.MarshalText()
		return string(b)
	}
	return ""
}

func (v *VariableLength) Bytes() []byte {
	if v != nil {
		return []byte(*v)
	}
	return []byte(nil)
}

func (p PublicKey) String() string {
	b, _ := p.MarshalText()
	return string(b)
}

func (p PublicKey) IsZero() bool {
	return p == zeroPublicKey
}

func (p *PublicKey) Bytes() []byte {
	if p != nil {
		return p[:]
	}
	return []byte(nil)
}

// Expects address in base58 form
func NewAccountFromAddress(s string) (*Account, error) {
	hash, err := crypto.NewRippleHashCheck(s, crypto.RIPPLE_ACCOUNT_ID)
	if err != nil {
		return nil, err
	}
	var account Account
	copy(account[:], hash.Payload())
	return &account, nil
}

func (a Account) Hash() (crypto.Hash, error) {
	return crypto.NewRippleAccount(a[:])
}

func (a Account) String() string {
	address, err := a.Hash()
	if err != nil {
		return fmt.Sprintf("Bad Address: %s", b2h(a[:]))
	}
	return address.String()
}

func (a Account) IsZero() bool {
	return a == zeroAccount
}

func (a *Account) Bytes() []byte {
	if a != nil {
		return a[:]
	}
	return []byte(nil)
}

func (a Account) Less(b Account) bool {
	return bytes.Compare(a[:], b[:]) < 0
}

func (a Account) Equals(b Account) bool {
	return a == b
}

func (a Account) Hash256() Hash256 {
	var h Hash256
	copy(h[:], a[:])
	return h
}

// Expects address in base58 form
func NewRegularKeyFromAddress(s string) (*RegularKey, error) {
	hash, err := crypto.NewRippleHashCheck(s, crypto.RIPPLE_ACCOUNT_ID)
	if err != nil {
		return nil, err
	}
	var regKey RegularKey
	copy(regKey[:], hash.Payload())
	return &regKey, nil
}

func (r RegularKey) Hash() (crypto.Hash, error) {
	return crypto.NewRippleAccount(r[:])
}

func (r RegularKey) String() string {
	address, err := r.Hash()
	if err != nil {
		return fmt.Sprintf("Bad Address: %s", b2h(r[:]))
	}
	return address.String()
}

func (r *RegularKey) Bytes() []byte {
	if r != nil {
		return r[:]
	}
	return []byte(nil)
}
