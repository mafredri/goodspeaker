/*

Package goodspeaker implements a reader and writer for communicating
with LG speakers (usually over TCP port 9741).

By providing the AES encryption key and IV, encrypted messages can also
be read and written. At least some speakers support plain-text
communication, in those cases the encryption key is not mandatory.

*/
package goodspeaker

import (
	"context"
	"crypto/aes"
	"crypto/cipher"

	"github.com/mafredri/goodspeaker/js/net"
)

// Header types.
const (
	headerPlainText = 0x0
	headerEncrypted = 0x10
)

type options struct {
	aes *aesBlock
}

// Option configures the reader and writer.
type Option func(*options)

type aesBlock struct {
	cipher.Block
	iv []byte
}

func newAESBlock(key, iv []byte) (*aesBlock, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &aesBlock{
		Block: b,
		iv:    iv,
	}, nil
}

func (b *aesBlock) newDecrypter() cipher.BlockMode {
	return cipher.NewCBCDecrypter(b.Block, b.iv)
}

func (b *aesBlock) newEncrypter() cipher.BlockMode {
	return cipher.NewCBCEncrypter(b.Block, b.iv)
}

// WithAES enables AES encrypted communication,
// requires a valid encryption key and IV.
func WithAES(key, iv []byte) (Option, error) {
	aes, err := newAESBlock(key, iv)
	if err != nil {
		return nil, err
	}

	return func(o *options) {
		o.aes = aes
	}, nil
}

// Dial connects to the speaker at the provided address.
//
// Building for js is supported.
func Dial(ctx context.Context, addr string) (net.Conn, error) {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func errWrap(err error, format string, v ...interface{}) error {
	if err == nil {
		return nil
	}
	if err == io.EOF { // Avoid wrapping sentinel io.EOF.
		return err
	}
	return errors.Errorf("%s: %w", fmt.Sprintf(format, v...), err)
}
