package ssh

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	gossh "golang.org/x/crypto/ssh"
)

// ssh.ParseSSHPublicKey parses an SSH public key in OpenSSH format
func ParseSSHPublicKey(keyBytes []byte) (*rsa.PublicKey, error) {
	// Parse the SSH public key
	sshPubKey, _, _, _, err := gossh.ParseAuthorizedKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH public key: %w", err)
	}

	// Convert to crypto.PublicKey
	cryptoPubKey, ok := sshPubKey.(gossh.CryptoPublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not a crypto public key")
	}

	// Convert to RSA public key
	rsaPubKey, ok := cryptoPubKey.CryptoPublicKey().(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not an RSA key")
	}

	return rsaPubKey, nil
}

// ParseSSHPrivateKey parses an SSH private key from PEM format
func ParseSSHPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}
