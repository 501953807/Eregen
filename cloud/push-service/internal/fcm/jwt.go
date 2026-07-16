package fcm

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"time"
)

// m is a shorthand for map literals used in JWT headers/payloads.
type m = map[string]interface{}

// jwtBearerGrant performs a JWT-Bearer exchange with Google OAuth2.
func jwtBearerGrant(ctx context.Context, cli *http.Client, privateKeyPEM, keyID, clientEmail string) (string, time.Time, error) {
	priv, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("parse private key: %w", err)
	}

	header := base64.RawURLEncoding.EncodeToString(jsonRaw(m{"alg": "RS256", "typ": "JWT", "kid": keyID}))
	payload := base64.RawURLEncoding.EncodeToString(jsonRaw(m{
		"iss": clientEmail, "sub": clientEmail,
		"aud": "https://oauth2.googleapis.com/token",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}))

	signingInput := header + "." + payload
	hashed := sha256.Sum256([]byte(signingInput))
	signature, err := priv.Sign(rand.Reader, hashed[:], nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}

	jwt := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)

	resp, err := cli.PostForm("https://oauth2.googleapis.com/token", map[string][]string{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jwt},
	})
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	data, _ := io.ReadAll(resp.Body)
	json.Unmarshal(data, &result)
	if result.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("fcm oauth error: %s", string(data))
	}

	return result.AccessToken, time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second), nil
}

func metadataAccessToken(ctx context.Context, cli *http.Client) (string, time.Time, error) {
	req, _ := http.NewRequest("GET",
		"http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token?scopes=https://www.googleapis.com/auth/cloud-platform", nil)
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := cli.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	data, _ := io.ReadAll(resp.Body)
	json.Unmarshal(data, &result)
	if result.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("metadata token error: %s", string(data))
	}
	return result.AccessToken, time.Now().Add(time.Duration(result.ExpiresIn-60)*time.Second), nil
}

func jsonRaw(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse pkcs1: %w", err)
		}
	}
	if sk, ok := key.(*rsa.PrivateKey); ok {
		return sk, nil
	}
	return nil, fmt.Errorf("not an RSA key")
}
