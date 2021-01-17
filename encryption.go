package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"

	//"encoding/gob"
	"encoding/pem"
)

// Encryption : This is the object that defines an encryption key class.
type Encryption struct {
	key     *rsa.PrivateKey
	bitSize int
}

// GenerateKey : Generate a set of keys.
func (enc *Encryption) GenerateKey() bool {
	// Create a new rand reader.
	reader := rand.Reader

	// Generate the key.
	key, err := rsa.GenerateKey(reader, enc.bitSize)

	// Return if there was an error.
	if err != nil {
		return false
	}

	// Save the key in the struct.
	enc.key = key

	return true
}

// GetPrivateKey : Get the PEM private key string.
func (enc *Encryption) GetPrivateKey() string {
	// Create the PEM private key block.
	privateKey := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(enc.key),
	}

	// Get the private key bytes.
	privateBytes := pem.EncodeToMemory(privateKey)

	// Return the string of the private bytes.
	return string(privateBytes)
}

// GetPublicKey : Get the public PEM key string.
func (enc *Encryption) GetPublicKey() string {
	// Get the asn1 bytes from the public key.
	asn1Bytes, err := asn1.Marshal(enc.key.PublicKey)

	if err != nil {
		return ""
	}

	// Create the PEM block.
	publicKey := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Encode the key into memory.
	strPubKey := pem.EncodeToMemory(publicKey)

	// Return the string representation.
	return string(strPubKey)
}
