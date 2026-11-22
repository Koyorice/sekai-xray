package proxy

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type Crypt struct {
	iv        []byte
	block     cipher.Block
	blockSize int
}

func NewCrypt(iv []byte, key []byte) *Crypt {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Errorf("crypt: invalid AES key: %w", err))
	}

	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		panic(fmt.Errorf("crypt: invalid IV length %d", len(iv)))
	}

	return &Crypt{
		iv:        iv,
		block:     block,
		blockSize: blockSize,
	}
}

func (c *Crypt) Encrypt(data []byte) ([]byte, error) {
	padLen := c.blockSize - len(data)%c.blockSize

	encrypted := make([]byte, len(data)+padLen)
	copy(encrypted, data)

	for i := len(data); i < len(encrypted); i++ {
		encrypted[i] = byte(padLen)
	}

	cipher.NewCBCEncrypter(c.block, c.iv).CryptBlocks(encrypted, encrypted)

	return encrypted, nil
}

func (c *Crypt) Decrypt(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data)%c.blockSize != 0 {
		return nil, fmt.Errorf("crypt: invalid ciphertext length %d", len(data))
	}

	decrypted := make([]byte, len(data))
	cipher.NewCBCDecrypter(c.block, c.iv).CryptBlocks(decrypted, data)

	length := len(decrypted)
	padLen := int(decrypted[length-1])

	if padLen <= 0 || padLen > length || padLen > c.blockSize {
		return nil, fmt.Errorf("crypt: invalid padding %d", padLen)
	}

	for i := length - padLen; i < length; i++ {
		if decrypted[i] != byte(padLen) {
			return nil, fmt.Errorf("crypt: invalid padding byte %d", decrypted[i])
		}
	}

	return decrypted[:length-padLen], nil
}
