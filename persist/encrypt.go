package persist

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"io"
	"os"
)

var key []byte

// EncryptData encrypts the given data and writes it to the specified file
func EncryptData(data interface{}, outputFile string) error {
	outFile, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer outFile.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	if _, err := outFile.Write(iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: outFile}

	encoder := gob.NewEncoder(writer)
	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

// DecryptData reads the encrypted data from the specified file, decrypts it, and populates the given data struct
func DecryptData(inputFile string, data interface{}) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(inFile, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	reader := &cipher.StreamReader{S: stream, R: inFile}

	decoder := gob.NewDecoder(reader)
	if err := decoder.Decode(data); err != nil {
		return err
	}

	return nil
}

func Initialize(bytes []byte) error {
	key = bytes

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.Mkdir(dir, 0755)
	}

	return nil
}
