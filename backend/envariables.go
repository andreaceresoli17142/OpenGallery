package main

import ( // {{{

	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
) // }}}

var databaseString string
var hostSite string
var sqlServerIp string
var dbname string

var clientId string
var clientSecret string
var redirectUri string

var email_email string
var email_password string
var email_server string
var email_port string

var rsaPrivateKey rsa.PrivateKey
var rsaPublicKey rsa.PublicKey

func loadEnv() bool {
	godotenv.Load(".env")

	hostSite, ok := os.LookupEnv("HOST_SITE")
	if !ok {
		Errorln("missing host site from env variables")
		return false
	}

	//o needs change
	redirectUri = hostSite + "oauth"

	sqlServerIp, ok = os.LookupEnv("SQL_SERVER_IP")
	if !ok {
		Errorln("missing sql server ip from env variables")
		return false
	}

	dbname, ok = os.LookupEnv("DATABASE_NAME")
	if !ok {
		Errorln("missing database name from env variables")
		return false
	}

	databaseString = "root:root@tcp(" + sqlServerIp + ")/" + dbname

	clientId, ok = os.LookupEnv("CLIENT_ID_OAUTH")
	if !ok {
		Errorln("missing client id from env variables")
		return false
	}

	clientSecret, ok = os.LookupEnv("CLIENT_SECRET_OAUTH")
	if !ok {
		Errorln("missing client secret from env variables")
		return false
	}

	email_email, ok = os.LookupEnv("EMAIL_EMAIL")
	if !ok {
		Errorln("missing noreply email from env variables")
		return false
	}

	email_password, ok = os.LookupEnv("EMAIL_PASSWORD")
	if !ok {
		Errorln("missing noreply password from env variables")
		return false
	}

	email_server, ok = os.LookupEnv("EMAIL_SERVER")
	if !ok {
		Errorln("missing noreply server from env variables")
		return false
	}

	email_port, ok = os.LookupEnv("EMAIL_PORT")
	if !ok {
		Errorln("missing noreply port from env variables")
		return false
	}

	// pk, ok := os.LookupEnv("PRIVATE_KEY")
	// if !ok {
	// 	// Errorln("missing noreply port from env variables")
	// 	// return false
	// 	kp, err := rsa.GenerateKey(rand.Reader, 2048)
	// 	if err != nil {
	// 		Errorln("error generating private key")
	// 		return false
	// 	}
	// 	rsaPrivateKey = *kp
	// 	pks := ExportRsaPrivateKeyAsPemStr(kp)
	// 	os.Setenv("PRIVATE_KEY", pks)
	// } else {
	// 	rpk, err := ParseRsaPrivateKeyFromPemStr(pk)
	// 	if err != nil {
	// 		Errorln("error parsing private key")
	// 		return false
	// 	}
	// 	rsaPrivateKey = *rpk
	// }

	// rsaPublicKey = rsaPrivateKey.PublicKey

	// secretMessage := "0"

	// encryptedMessage, err := rsa_Encrypt(secretMessage, rsaPublicKey)
	// if err != nil {
	// 	Errorln("error parsing private key")
	// 	return false
	// }

	// plainText, err := rsa_Decrypt(encryptedMessage, rsaPrivateKey)

	// if plainText != secretMessage || err != nil {
	// 	Errorln("error generating key pais")
	// 	return false
	// }

	// Successln("created key pairs")

	Successln("enviroment variables loaded")
	return true
}

func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func rsa_Encrypt(secretMessage string, key rsa.PublicKey) (string, error) {
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &key, []byte(secretMessage), label)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func rsa_Decrypt(cipherText string, privKey rsa.PrivateKey) (string, error) {
	ct, _ := base64.StdEncoding.DecodeString(cipherText)
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, &privKey, ct, label)
	if err != nil {
		return "", err
	}
	// fmt.Println("Plaintext:", string(plaintext))
	return string(plaintext), nil
}

func verifyRsaSignature(publicKey *rsa.PublicKey, keyString string) (string, error) {

	dataSigPair := strings.Split(keyString, ".")

	getTkDigest := sha256.Sum256([]byte(dataSigPair[0]))

	decodedTk, _ := base64.StdEncoding.DecodeString(dataSigPair[1])

	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, getTkDigest[:], decodedTk)

	if err != nil {
		return "", err
	}
	return dataSigPair[0], nil
}

func generateRsaSignature(msg []byte, pk *rsa.PrivateKey) (string, error) {

	msgDigest := sha256.Sum256(msg)

	signature, err := rsa.SignPKCS1v15(rand.Reader, pk, crypto.SHA256, msgDigest[:])

	if err != nil {
		return "", AppendError("error generating signture", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}
