package util

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func CreatePrivateKeyPem(filePath string, priv *rsa.PrivateKey) error {
	keyOut, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, pemBlockForKey(priv))
	keyOut.Close()
	return nil
}

func ReadPrivateKeyPem(filePath string) (*rsa.PrivateKey, error) {
	pubPEMData, err := ioutil.ReadFile(filePath)
	block, _ := pem.Decode(pubPEMData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	prv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return prv, nil
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case crypto.PublicKey:
		fmt.Println("public key")
		asn1Bytes, err := x509.MarshalPKIXPublicKey(k)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return &pem.Block{Type: "RSA PUBLIC KEY", Bytes: asn1Bytes}
	// case *ecdsa.PrivateKey:
	// 	b, err := x509.MarshalECPrivateKey(k)
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
	// 		os.Exit(2)
	// 	}
	// 	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func EncodePublicKeyPem(publicKey crypto.PublicKey) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	pemBlock := pemBlockForKey(publicKey)
	if pemBlock == nil {
		return nil, errors.New("Can not generate pem.")
	}
	pem.Encode(buf, pemBlock)

	return buf, nil
}

func EncodePublicKey(publicKey crypto.PublicKey) (*bytes.Buffer, error) {
	asn1Bytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	encoder.Write(asn1Bytes)
	encoder.Close()
	return buf, err
}

func DecodePublicKey(base64Key string) (*rsa.PublicKey, error) {
	decodedata, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(decodedata)
	if err != nil {
		return nil, errors.New("failed to parse DER encoded public key: " + err.Error())
	}
	rsa_pub, ok := pub.(*rsa.PublicKey)
	if ok {
		return rsa_pub, nil
	} else {
		return nil, errors.New("cant not conver to rsa_pub")
	}
}

func EncodeString(str string, publicKey *rsa.PublicKey) (*bytes.Buffer, error) {
	encryptPwd, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(str))
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	encoder.Write(encryptPwd)
	encoder.Close()
	return buf, nil
}

func DecodeString(encodeStr string, privateKey *rsa.PrivateKey) (*bytes.Buffer, error) {
	decodeStr, err := base64.StdEncoding.DecodeString(encodeStr)
	if err != nil {
		return nil, err
	}

	plaintextByte, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, decodeStr)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(plaintextByte), nil
}

func EncodeMap(dataMap *map[string]interface{}) (string, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(dataMap)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func DecodeMap(ecodeMapStr string) (*map[string]interface{}, error) {
	mapByte, err := base64.StdEncoding.DecodeString(ecodeMapStr)
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	b.Write(mapByte)
	d := gob.NewDecoder(&b)
	result := make(map[string]interface{})
	err = d.Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func MD5(str string) string {
	w := md5.New()
	io.WriteString(w, str)
	return hex.EncodeToString(w.Sum(nil))
}

func MD5Byte(data []byte) string {
	w := md5.New()
	w.Write(data)
	return hex.EncodeToString(w.Sum(nil))
}

func SHA1(str string) string {
	h := sha1.New()
	io.WriteString(h, "His money is twice tainted:")
	io.WriteString(h, " 'taint yours and 'taint mine.")
	return hex.EncodeToString(h.Sum(nil))
}
