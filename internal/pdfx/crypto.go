package pdfx

import (
	"context"
	"errors"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Protect encrypts inFile with AES-256. The password is needed to open the
// result and to change it (user and owner password are the same).
func Protect(ctx context.Context, inFile, outFile, password string) error {
	if password == "" {
		return errors.New("password is empty")
	}
	c := conf()
	c.UserPW, c.OwnerPW = password, password
	c.EncryptUsingAES = true
	c.EncryptKeyLength = 256
	return api.EncryptFile(ctx, inFile, outFile, c)
}

// Unlock removes the password from inFile.
func Unlock(ctx context.Context, inFile, outFile, password string) error {
	c := conf()
	c.UserPW, c.OwnerPW = password, password
	return api.DecryptFile(ctx, inFile, outFile, c)
}
