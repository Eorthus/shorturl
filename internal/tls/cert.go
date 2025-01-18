// Package tls предоставляет функции для работы с TLS-сертификатами
package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

// GenerateSelfSignedCert генерирует самоподписанный сертификат и приватный ключ
func GenerateSelfSignedCert(certFile, keyFile string) error {
	// Создаем приватный RSA-ключ длиной 4096 бит
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// Создаем шаблон сертификата
	template := x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"URL Shortener Self Signed"},
			Country:      []string{"RU"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0), // Действителен 10 лет
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		// Разрешаем использование для localhost
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	// Создаем сертификат x509
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// Сохраняем сертификат в PEM-формате
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	// Сохраняем приватный ключ в PEM-формате
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return err
	}

	return nil
}

// EnsureCertificateExists проверяет наличие сертификата и ключа, создает их если отсутствуют
func EnsureCertificateExists(certFile, keyFile string) error {
	_, certErr := os.Stat(certFile)
	_, keyErr := os.Stat(keyFile)

	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		return GenerateSelfSignedCert(certFile, keyFile)
	}

	return nil
}
