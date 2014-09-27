package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"net/http"
	"net/url"
	"os"

	"github.com/go-martini/martini"
	"github.com/looking-promising/privfiles/server/crypto"
	"github.com/looking-promising/privfiles/server/model"
)

type FileController struct {
	masterKey []byte
}

func (ctlr *FileController) Upload(res http.ResponseWriter, req *http.Request) string {

	// get a mulitpart reader from the request
	mr, err := req.MultipartReader()
	if err != nil {
		// fmt.Println(err)
		return err.Error()
	}

	// generate the encryption key
	key := crypto.GenerateKey(32)

	// encrypt the file
	fileId, err := crypto.EncryptMultipartReader(mr, req.ContentLength, key)

	//fmt.Println("fileName: ", fileName)
	//fmt.Println("contentType:", contentType)

	// marshal the fileId to a []byte of json data
	marshaledFileId, err := json.Marshal(fileId)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}

	// Convert []byte to string
	serializedFileId := base64.URLEncoding.EncodeToString(marshaledFileId)

	// build the url to be returned making sure it gets encoded
	urlToMAC := ctlr.buildDownloadUrl(serializedFileId)
	mac := crypto.ComputeMAC(urlToMAC, ctlr.masterKey)

	//fmt.Printf("URL MAC: %q\n", mac)

	var urlWithMAC *url.URL
	urlWithMAC, err = urlWithMAC.Parse(urlToMAC + "/" + mac)
	if err != nil {
		//panic("server error")
		return err.Error()
	}

	//fmt.Printf("Encoded URL is %q\n", urlWithMAC.String())

	// return the json result containing the download URL
	return "{\"url\":\"" + urlWithMAC.String() + "\"}"
}

func (ctlr *FileController) Download(params martini.Params, res http.ResponseWriter, req *http.Request) {
	serializedFileId := params["fileId"]
	mac := params["mac"]

	//fmt.Println("serializedFileId: ", serializedFileId)
	//fmt.Println("mac: ", mac)

	message := ctlr.buildDownloadUrl(serializedFileId)
	if !crypto.CheckMAC(message, mac, ctlr.masterKey) {
		panic("FAILURE! URL may have been corrupted or tampered with.")
	}

	jsonToUnmarshal, err := base64.URLEncoding.DecodeString(serializedFileId)
	if err != nil {
		panic("server error")
	}

	var fileId model.FileIdentifier
	err = json.Unmarshal(jsonToUnmarshal, &fileId)
	if err != nil {
		panic("server error")
	}

	//fmt.Println("fileId: ", fileId)
	file := fileId.Path

	// decode the key back to a []byte
	//key, err := base64.URLEncoding.DecodeString(fileId.Key)
	//if err != nil {
	//	panic("server error")
	//}

	//set the relevant headers.
	res.Header().Set("Content-Disposition", "attachment; filename="+fileId.FileName)
	res.Header().Set("Content-Type", fileId.ContentType)

	err = crypto.DecryptStream(res, file, fileId.Key, fileId.Checksum)
	if err != nil {
		panic(err.Error())
	}

	err = os.Remove(file)

	if err != nil {
		//fmt.Println(err)
		//panic(err)
		return
	}

	return
}

func (ctlr *FileController) buildDownloadUrl(serializedFileId string) string {
	// build the url to be returned making sure it gets encoded
	var baseUrl *url.URL
	baseUrl, err := baseUrl.Parse("http://privfiles.com:3000/download/" + serializedFileId)
	if err != nil {
		panic("server error")
	}

	return (baseUrl.String())
}
