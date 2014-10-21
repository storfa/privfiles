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
	fileIdGrpoup, err := crypto.EncryptMultipartReader(mr, req.ContentLength, key)

	// marshal the fileIdGroup to a []byte of json data
	marshaledFileIdGroup, err := json.Marshal(fileIdGroup)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}

	// Convert []byte to string
	serializedFileIdGroup := base64.URLEncoding.EncodeToString(marshaledFileIdGroup)

	// build the url to be returned making sure it gets encoded
	urlToMAC := ctlr.buildDownloadUrl(serializedFileIdGroup)
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
	serializedFileIdGroup := params["fileId"]
	mac := params["mac"]

	//fmt.Println("serializedFileIdGroup: ", serializedFileIdGroup)
	//fmt.Println("mac: ", mac)

	message := ctlr.buildDownloadUrl(serializedFileIdGroup)
	if !crypto.CheckMAC(message, mac, ctlr.masterKey) {
		panic("FAILURE! URL may have been corrupted or tampered with.")
	}

	jsonToUnmarshal, err := base64.URLEncoding.DecodeString(serializedFileIdGroup)
	if err != nil {
		panic("server error")
	}

	var fileIdGroup model.FileIdentifierGroup
	err = json.Unmarshal(jsonToUnmarshal, &fileIdGroup)
	if err != nil {
		panic("server error")
	}

	//fmt.Println("fileIdGroup: ", fileIdGroup)

	if len(fileIdGroup.FileIds) == 1 {
		file := fileIdGroup.Path

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
	} else if len(fileIdGroup.FileIds) > 1 {
		// TODO:  Allow user to download all files as a zip if there is more than one.
		//	      otherwise, send the single file to the user with the correct ContentType.
		fmt.Println("blah")
	} else {
		//TODO: return error
		fmt.Println("blah")
	}

	return
}

func (ctlr *FileController) buildDownloadUrl(serializedFileIdGroup string) string {
	// build the url to be returned making sure it gets encoded
	var baseUrl *url.URL
	baseUrl, err := baseUrl.Parse("http://privfiles.com:3000/download/" + serializedFileIdGroup)
	if err != nil {
		panic("server error")
	}

	return (baseUrl.String())
}
