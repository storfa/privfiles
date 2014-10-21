package model

type FileIdentifier struct {
	Checksum    string
	FileName    string
	StoredName  string
	ContentType string
}

type FileIdentifierGroup struct {
	Key       []byte
	GroupPath string
	FileIds   []FileIdentifier
}
