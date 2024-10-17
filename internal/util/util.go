package util

import "golang.org/x/crypto/bcrypt"

func GetHashPassword(password string) ([]byte, error) {
	bytePassword := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hash, nil
}
