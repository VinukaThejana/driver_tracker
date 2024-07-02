// Package lib contains various library packages
package lib

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/VinukaThejana/go-utils/logger"
	"github.com/bytedance/sonic"
)

// LogFatal is a function that is used to Run various functions that need to crash if an error is found
func LogFatal(err error) {
	if err != nil {
		logger.Errorf(err)
	}
}

// ToStr is a function that is used to convert bad json strings to nicer json bytes
func ToStr(jsonStr string) ([]byte, error) {
	var (
		data    map[string]interface{}
		payload []byte
		err     error
	)

	if err = sonic.UnmarshalString(jsonStr, &data); err != nil {
		return nil, err
	}
	if payload, err = sonic.Marshal(data); err != nil {
		return nil, err
	}

	return payload, nil
}

type response map[string]interface{}

// JSONResponse is a function that is used to send a a simple JSON response to the client
func JSONResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	sonic.ConfigDefault.NewEncoder(w).Encode(response{
		"message": message,
	})
}

// JSONResponseWInterface is a function to send a JSON response with the given interface
func JSONResponseWInterface(w http.ResponseWriter, statusCode int, res map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	sonic.ConfigDefault.NewEncoder(w).Encode(res)
}

// Base64URLDecode decodes base64url string to byte array
// The following functions are adapted from the code provided by DV[dvsekhvalnov]
// Source: https://github.com/dvsekhvalnov/jose2go/blob/master/base64url/base64url.go
func Base64URLDecode(data string) ([]byte, error) {
	data = strings.ReplaceAll(data, "-", "+") // 62nd char of encoding
	data = strings.ReplaceAll(data, "_", "/") // 63rd char of encoding

	switch len(data) % 4 { // Pad with trailing '='s
	case 0: // no padding
	case 2:
		data += "==" // 2 pad chars
	case 3:
		data += "=" // 1 pad char
	}

	return base64.StdEncoding.DecodeString(data)
}

// Base64URLEncode encodes given byte array to base64url string
// The following functions are adapted from the code provided by DV[dvsekhvalnov]
// Source: https://github.com/dvsekhvalnov/jose2go/blob/master/base64url/base64url.go
func Base64URLEncode(data []byte) string {
	result := base64.StdEncoding.EncodeToString(data)
	result = strings.ReplaceAll(result, "+", "-") // 62nd char of encoding
	result = strings.ReplaceAll(result, "/", "_") // 63rd char of encoding
	result = strings.ReplaceAll(result, "=", "")  // Remove any trailing '='s

	return result
}

// GenerateToken is a function that is used to generate a token of a given length
func GenerateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
