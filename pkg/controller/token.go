package controller

import (
	"golang.org/x/crypto/bcrypt"
	"encoding/base64"
	"github.com/golang/glog"
)

// GenerateToken returns a unique token based on the provided string
func GenerateToken(name string, cost int) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(name), cost)
	if err != nil {
		glog.Errorf("Failed to generate a random token: %v", err)
	}
	return base64.StdEncoding.EncodeToString(hash)
}
