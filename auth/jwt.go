package auth

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type JwtToken interface {
	GetTokenWithoutExpired(host string, data map[string]interface{}) (*string, error)
	GetToken(host string, data map[string]interface{}, exp uint8) (*string, error)
	ParseToken(tokenStr string) (*jwt.Token, error)
	ParseTokenUnValidate(tokenStr string) (*jwt.Token, error)
	// 對特定資源存取金鑰
	GetJwtAccessToken(host string, source string, id interface{}, db string, perm UserPerm) (*string, error)
	GetCompanyToken(host, compID, compName, userID, acc, userName string, perm UserPerm) (*string, error)
}

type JwtDI interface {
	GetKid() string
	NewJwt() JwtToken
}

type JwtConf struct {
	PrivateKeyFile string `yaml:"privatekey"`
	PublicKeyFile  string `yaml:"publickey"`
	Header         struct {
		Alg string `yaml:"alg"`
		Typ string `yaml:"typ"`
		Kid string `yaml:"kid"`
	} `yaml:"header"`
	Claims struct {
		ExpDuration time.Duration `yaml:"exp"`
	} `yaml:"claims"`

	myHeader   map[string]interface{}
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func (j *JwtConf) getHeader() map[string]interface{} {
	if j.myHeader != nil {
		delete(j.myHeader, "usa")
		return j.myHeader
	}
	j.myHeader = map[string]interface{}{
		"alg": j.Header.Alg,
		"typ": j.Header.Typ,
		"kid": j.Header.Kid,
	}
	return j.myHeader
}

func (j *JwtConf) getPublicKey() (*rsa.PublicKey, error) {
	if j.publicKey != nil {
		return j.publicKey, nil
	}
	publicData, err := ioutil.ReadFile(j.PublicKeyFile)
	if err != nil {
		return nil, err
	}
	j.publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicData)
	return j.publicKey, err
}

func (j *JwtConf) getPrivateKey() (*rsa.PrivateKey, error) {
	if j.privateKey != nil {
		return j.privateKey, nil
	}
	privateData, err := ioutil.ReadFile(j.PrivateKeyFile)
	if err != nil {
		return nil, err
	}
	j.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateData)
	return j.privateKey, err
}

func (j *JwtConf) GetKid() string {
	return j.Header.Kid
}

func (j *JwtConf) NewJwt() JwtToken {
	return j
}

func (j *JwtConf) ParseTokenUnValidate(tokenStr string) (*jwt.Token, error) {
	if j == nil {
		return nil, errors.New("jwtConf is nil")
	}
	parser := jwt.Parser{
		SkipClaimsValidation: true,
	}

	token, err := parser.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		pk, err := j.getPublicKey()
		return pk, err
	})
	if token == nil {
		return nil, errors.New("token is nil")
	}
	if token.Valid {
		return token, nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, errors.New("That's not even a token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			return nil, errors.New("Timing is everything")
		} else {
			return nil, err
		}
	}
	return nil, err
}
func (j *JwtConf) ParseToken(tokenStr string) (*jwt.Token, error) {
	if j == nil {
		return nil, errors.New("jwtConf is nil")
	}
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		pk, err := j.getPublicKey()
		return pk, err
	})
	if token == nil {
		return nil, errors.New("token is nil")
	}
	if token.Valid {
		return token, nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, errors.New("That's not even a token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			return nil, errors.New("Timing is everything")
		} else {
			return nil, err
		}
	}
	return nil, err
}

func (j *JwtConf) GetToken(host string, data map[string]interface{}, exp uint8) (*string, error) {
	if j == nil {
		return nil, errors.New("jwtConf not set")
	}
	if data == nil {
		return nil, errors.New("no data")
	}
	if exp <= 0 {
		exp = 60
	} else if exp > 180 {
		exp = 180
	}

	now := time.Now()
	data["iss"] = host
	data["iat"] = now.Unix()
	if exp > 0 {
		data["exp"] = now.Add(time.Duration(exp) * time.Minute).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(data))

	token.Header = j.getHeader()

	pk, err := j.getPrivateKey()
	if err != nil {
		return nil, err
	}
	ss, err := token.SignedString(pk)
	if err != nil {
		return nil, err
	}
	return &ss, nil
}

func (j *JwtConf) GetTokenWithoutExpired(host string, data map[string]interface{}) (*string, error) {
	if j == nil {
		return nil, errors.New("jwtConf not set")
	}
	if data == nil {
		return nil, errors.New("no data")
	}

	now := time.Now()
	data["iss"] = host
	data["iat"] = now.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(data))

	token.Header = j.getHeader()

	pk, err := j.getPrivateKey()
	if err != nil {
		return nil, err
	}
	ss, err := token.SignedString(pk)
	if err != nil {
		return nil, err
	}
	return &ss, nil
}

func (j *JwtConf) GetJwtAccessToken(host string, source string, id interface{}, db string, perm UserPerm) (*string, error) {
	if j == nil {
		return nil, errors.New("jwtConf not set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(map[string]interface{}{
		"iss":      host,
		"source":   source,
		"sourceId": id,
		"db":       db,
		"per":      perm,
	}))

	token.Header = j.getHeader()
	token.Header["usa"] = "access"

	pk, err := j.getPrivateKey()
	if err != nil {
		return nil, err
	}
	ss, err := token.SignedString(pk)
	if err != nil {
		return nil, err
	}
	return &ss, nil
}

func (j *JwtConf) GetCompanyToken(host, compID, compName, userID, acc, userName string, perm UserPerm) (*string, error) {
	if j == nil {
		return nil, errors.New("jwtConf not set")
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(map[string]interface{}{
		"iss":    host,
		"iat":    now.Unix(),
		"exp":    now.Add(time.Duration(180) * time.Minute).Unix(),
		"comp":   compName,
		"compID": compID,
		"per":    perm,
		"sub":    userID,
		"acc":    acc,
		"nam":    userName,
	}))

	token.Header = j.getHeader()
	token.Header["usa"] = "comp"

	pk, err := j.getPrivateKey()
	if err != nil {
		return nil, err
	}
	ss, err := token.SignedString(pk)
	if err != nil {
		return nil, err
	}
	return &ss, nil
}
