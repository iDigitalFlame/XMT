// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package data

import (
	"crypto/elliptic"
	"crypto/rand"
	"io"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	sharedKeySize  = 65
	publicKeySize  = 133
	privateKeySize = 66
)

var keyCurve = elliptic.P521()

// KeyPair is a ECDH key pair that can be used to generate and manage Public,
// Private and shared keys. The data from this struct can be written and read
// from disk or network.
//
// The empty value can be filled using the 'Fill*' functions.
//
// If not changed via the 'keyCurve' const, this function uses NIST P-521 (FIPS
// 186-3, section D.2.5), known as secp521r1.
//
// Initial ideas and concepts from: https://github.com/wsddn/go-ecdh
type KeyPair struct {
	Public  PublicKey
	Private PrivateKey
	share   SharedKeys
}

// PublicKey represents a ECDH PublicKey in raw binary format. This alias can be
// used to parse or output to a string value.
type PublicKey [publicKeySize]byte

// PrivateKey represents a ECDH PrivateKey in raw binary format. This alias can
// be used to parse or output to a string value.
type PrivateKey [privateKeySize]byte

// SharedKeys represents a ECDH SharedKey in raw binary format. This alias is just
// used to differentate the shared key from other binary structures.
type SharedKeys [sharedKeySize]byte

// Fill will populate this KeyPair with a randomally generated Public and Private
// key values.
//
// Before returning, this function will zero out the shared secret.
func (k *KeyPair) Fill() {
	// We can ignore the error, since it's only for the read operation.
	var (
		p, x, y, _ = elliptic.GenerateKey(keyCurve, rand.Reader)
		s          = elliptic.Marshal(keyCurve, x, y)
	)
	copy(k.Public[:], s)
	copy(k.Private[:], p)
	p, x, y, s = nil, nil, nil, nil
	for i := range k.share {
		k.share[i] = 0
	}
}

// Empty returns true if the PublicKey is empty (all zeros).
func (k *KeyPair) Empty() bool {
	return k.Public.Empty()
}

// Sync attempts to generate the Shared key using the current KeyPair's Public
// and Private key values.
//
// This function returns an error if a Shared key could not be generated.
func (k *KeyPair) Sync() error {
	return k.fillShared(k.Public, k.Private)
}

// Empty returns true if this PublicKey is empty (all zeros).
func (p PublicKey) Empty() bool {
	for i := range p {
		if p[i] > 0 {
			return false
		}
	}
	return true
}

// Hash returns the FNV-32 hash of this PublicKey in a uint32 format.
func (p PublicKey) Hash() uint32 {
	h := uint32(2166136261)
	for i := range p {
		h *= 16777619
		h ^= uint32(p[i])
	}
	return h
}

// IsSynced returns false if the Shared key is empty (all zeros).
func (k *KeyPair) IsSynced() bool {
	for i := range k.share {
		if k.share[i] > 0 {
			return true
		}
	}
	return false
}

// Shared returns a copy of the current Shared key contained in this KeyPair.
//
// If 'IsSynced' returns false, the output will be a zero filled array.
func (k KeyPair) Shared() SharedKeys {
	return k.share
}

// Read will read in the PublicKey ONLY from the supplied 'io.Reader' and fill
// the current KeyPair's PublicKey with the resulting data.
//
// Any errors or invalid byte lengths read will return an error.
func (k *KeyPair) Read(r io.Reader) error {
	switch n, err := r.Read(k.Public[:]); {
	case err != nil:
		return err
	case n != publicKeySize:
		return io.ErrUnexpectedEOF
	}
	return nil
}

// Write will write out the PublicKey ONLY to the supplied 'io.Writer'.
//
// Any errors or invalid byte lengths written will return an error.
func (k *KeyPair) Write(w io.Writer) error {
	switch n, err := w.Write(k.Public[:]); {
	case err != nil:
		return err
	case n != publicKeySize:
		return io.ErrShortWrite
	}
	return nil
}

// Marshal will write out the Public, Private and Shared key data to the supplied
// 'io.Writer'.
//
// Any errors or invalid byte lengths written will return an error.
func (k *KeyPair) Marshal(w io.Writer) error {
	switch n, err := w.Write(k.Public[:]); {
	case err != nil:
		return err
	case n != publicKeySize:
		return io.ErrShortWrite
	}
	switch n, err := w.Write(k.Private[:]); {
	case err != nil:
		return err
	case n != privateKeySize:
		return io.ErrShortWrite
	}
	switch n, err := w.Write(k.share[:]); {
	case err != nil:
		return err
	case n != sharedKeySize:
		return io.ErrShortWrite
	}
	return nil
}

// Unmarshal will read in the Public, Private and Shared key data from the supplied
// 'io.Reader' and fill all the current KeyPair data with the resulting data.
//
// Any errors or invalid byte lengths read will return an error.
func (k *KeyPair) Unmarshal(r io.Reader) error {
	switch n, err := r.Read(k.Public[:]); {
	case err != nil:
		return err
	case n != publicKeySize:
		return io.ErrUnexpectedEOF
	}
	switch n, err := r.Read(k.Private[:]); {
	case err != nil:
		return err
	case n != privateKeySize:
		return io.ErrUnexpectedEOF
	}
	switch n, err := r.Read(k.share[:]); {
	case err != nil:
		return err
	case n != sharedKeySize:
		return io.ErrUnexpectedEOF
	}
	return nil
}

// FillPublic will generate the Shared key using the KeyPair's PrivateKey and
// the supplied PublicKey.
//
// If successful, the PublicKey data will be copied over the current KeyPair's
// PublicKey for successive calls to 'Sync'.
//
// This function returns an error if a Shared key could not be generated.
func (k *KeyPair) FillPublic(p PublicKey) error {
	if err := k.fillShared(p, k.Private); err != nil {
		return err
	}
	copy(k.Public[:], p[:])
	return nil
}

// FillPrivate will generate the Shared key using the KeyPair's PublicKey and
// the supplied PrivateKey.
//
// If successful, the PrivateKey data will be copied over the current KeyPair's
// PrivateKey for successive calls to 'Sync'.
//
// This function returns an error if a Shared key could not be generated.
func (k *KeyPair) FillPrivate(p PrivateKey) error {
	if err := k.fillShared(k.Public, p); err != nil {
		return err
	}
	copy(k.Private[:], p[:])
	return nil
}
func (k *KeyPair) fillShared(n PublicKey, m PrivateKey) error {
	x, y := elliptic.Unmarshal(keyCurve, n[:])
	if x == nil || y == nil {
		return xerr.Sub("cannot parse curve PublicKey", 0x77)
	}
	v, _ := keyCurve.ScalarMult(x, y, m[:])
	if x, y = nil, nil; v == nil {
		return xerr.Sub("cannot multiply PrivateKey with PublicKey", 0x78)
	}
	copy(k.share[:], v.Bytes())
	x = nil
	return nil
}
