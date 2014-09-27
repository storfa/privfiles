package model

type FileIdentifier struct {
	Key         []byte
	Path        string
	Checksum    string
	FileName    string
	ContentType string
}
