package cryptokeys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/pkg/errors"
)

func GetPublicKey(cryptoKey string) (*rsa.PublicKey, error) {
	if cryptoKey == "" {
		return nil, nil
	}

	certificateBytes, err := os.ReadFile(cryptoKey)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось прочитать файл с публичным ключом")
	}

	pemBlock, _ := pem.Decode(certificateBytes)
	if pemBlock == nil {
		return nil, errors.New("не удалось декодировать pem-блок сертификата")
	}

	certificate, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить x509 сертификат")
	}

	publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("публичный ключ в сертификате не является rsa ключом")
	}

	return publicKey, nil
}

func GetPrivateKey(cryptoKey string) (*rsa.PrivateKey, error) {
	if cryptoKey == "" {
		return nil, nil
	}

	privateKeyBytes, err := os.ReadFile(cryptoKey)
	if err != nil {
		return nil, errors.New("не удалось прочитать файл с приватным ключом")
	}

	privateKeyPemBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyPemBlock == nil {
		return nil, errors.New("не удалось декодировать pem-блок приватного ключа")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes)
	if err != nil {
		return nil, errors.New("не удалось распарсить rsa приватный ключ")
	}

	return privateKey, nil
}
