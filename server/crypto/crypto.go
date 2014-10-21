package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"regexp"

	"github.com/looking-promising/privfiles/server/model"
)

func GenerateKey(numberOfBytes int) []byte {
	// define an slice of bytes
	b := make([]byte, numberOfBytes)

	// fill the byte slice with random data
	_, err := rand.Read(b)

	// check for errors
	if err != nil {
		panic(err)
	}

	return b
}

func ComputeMAC(message string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message))

	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

func CheckMAC(message string, messageMAC string, key []byte) bool {
	expectedMAC := ComputeMAC(message, key)
	return (messageMAC == expectedMAC)
}

func ComputeChecksum(filePath string) string {

	f, err := os.Open(filePath)

	if err != nil {
		panic(err)
	}
	defer f.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, f)

	if err != nil {
		panic(err)
	}

	return (base64.URLEncoding.EncodeToString(hasher.Sum(nil)))
}

func EncryptMultipartReader(mr *multipart.Reader, length int64, key []byte) (fileIdGroup model.FileIdentifierGroup, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	// a return value
	var tmpfilename string
	var tmpcontentType string

	// need a place to store the uploaded file
	fileIdGroup.Key = key
	fileIdGroup.FileIds = make([]model.FileIdentifier, 10)
	fileIdGroup.GroupPath = "/tmp/uploaded/" + base64.URLEncoding.EncodeToString(GenerateKey(8))

	for i := 0; true; i++ {
		part, e := mr.NextPart()
		if e == io.EOF {
			err = e
			break
		}

		// ****************************************
		// * TODO: extract this functionality
		// ****************************************

		// setup output file
		fileId := fileIdGroup.FileIds[i]
		fileId.StoredName = base64.URLEncoding.EncodeToString(GenerateKey(4))
		fileId.ContentType = part.Header.Get("Content-Type")

		// determine the file name
		dispositionHeader := part.Header.Get("Content-Disposition")
		re := regexp.MustCompile("(filename=\")(.*)(\")")
		fileNameSlices := re.FindStringSubmatch(dispositionHeader)
		fileId.FileName = fileNameSlices[2]

		destFilePath := fileIdGroup.GroupPath + "/" + fileId.StoredName
		outFile, e := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if e != nil {
			err = e
			return
		}
		defer outFile.Close()

		// Copy the input file to the output file, encrypting as we go.
		encryptor := &cipher.StreamWriter{S: stream, W: outFile}
		defer encryptor.Close()

		compressor, e := gzip.NewWriterLevel(encryptor, 1)
		if err != nil {
			err = e
			return
		}
		defer compressor.Close()
		// END: setup output file

		var read int64
		var p float32
		for {
			buffer := make([]byte, 100000)
			cBytes, err := part.Read(buffer)
			if err == io.EOF {
				break
			}

			read = read + int64(cBytes)
			p = float32(read) / float32(length) * 100

			log.Printf("progress: %v \n", p)

			if _, cerr := io.Copy(compressor, bytes.NewReader(buffer[0:cBytes])); cerr != nil {
				panic(err)
			}
		}

		fileId.Checksum = ComputeChecksum(destFilePath)

	}

	return
}

func DecryptStream(outStream io.Writer, filePath string, key []byte, checksum string) error {
	// TODO: figure out why checksums do not match (and they really don't)!!!
	fmt.Println("Provided Checksum: ", checksum)
	fmt.Println("Computed Checksum: ", ComputeChecksum(filePath))
	//	if computeChecksum(filePath) != checksum {
	//		return (errors.New("checksum does not match requested file indicating possible corruption or tampering."))
	//	}

	inFile, err := os.Open(filePath)
	if err != nil {
		return (err)
	}
	defer inFile.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return (err)
	}

	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	decryptor := &cipher.StreamReader{S: stream, R: inFile}
	decompressor, err := gzip.NewReader(decryptor)
	if err != nil {
		return (err)
	}
	defer decompressor.Close()

	// Copy the input file to the output stream, decrypting as we go.
	if _, err := io.Copy(outStream, decompressor); err != nil {
		return (err)
	}

	return (nil)
}

// // Encrypt applies the necessary padding to the message and encrypts it
// // with AES-CBC.
// func Encrypt(k, in []byte) ([]byte, bool) {
//     in = Pad(in)
//     iv := GenerateIV()
//     if iv == nil {
//         return nil, false
//     }
//
//     c, err := aes.NewCipher(k)
//     if err != nil {
//         return nil, false
//     }
//
//     cbc := cipher.NewCBCEncrypter(c, iv)
//     cbc.CryptBlocks(in, in)
//     return append(iv, in...), true
// }
//
//
// // Decrypt decrypts the message and removes any padding.
// func Decrypt(k, in []byte) ([]byte, bool) {
//     if len(in) == 0 || len(in)%aes.BlockSize != 0 {
//         return nil, false
//     }
//
//     c, err := aes.NewCipher(k)
//     if err != nil {
//         return nil, false
//     }
//
//     cbc := cipher.NewCBCDecrypter(c, in[:aes.BlockSize])
//     cbc.CryptBlocks(in[aes.BlockSize:], in[aes.BlockSize:])
//     out := Unpad(in[aes.BlockSize:])
//     if out == nil {
//         return nil, false
//     }
//     return out, true
//
// }
