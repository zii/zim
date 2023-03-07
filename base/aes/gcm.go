package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func GcmEncrypt(plaintext []byte, key, nonce []byte) ([]byte, error) {
	var aesgcm cipher.AEAD
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nil
}

func GcmEncryptString(plaintext string, key, nonce string) (string, error) {
	ciphertext, err := GcmEncrypt([]byte(plaintext), []byte(key), []byte(nonce))
	if err != nil {
		return "", err
	}
	out := base64.StdEncoding.EncodeToString(ciphertext)
	return out, nil
}

func GcmDecrypt(cipherdata []byte, key, nonce []byte) ([]byte, error) {
	var aesgcm cipher.AEAD
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesgcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aesgcm.Open(nil, []byte(nonce), cipherdata, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func GcmDecryptString(ciphertext string, key, nonce string) (string, error) {
	cipherdata, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	plaintext, err := GcmDecrypt(cipherdata, []byte(key), []byte(nonce))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
