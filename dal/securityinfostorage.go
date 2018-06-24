package dal

import (
	"encoding/xml"
	"os"

	"github.com/ChizhovVadim/assets/core"
)

type securityInfoStorage struct {
	path string
}

func NewSecurityInfoStorage(path string) *securityInfoStorage {
	return &securityInfoStorage{path}
}

func (srv *securityInfoStorage) ReadAll() ([]core.SecurityInfo, error) {
	var obj = struct {
		Items []core.SecurityInfo `xml:"SecurityInfo"`
	}{}
	var err = decodeXmlFile(srv.path, &obj)
	if err != nil {
		return nil, err
	}
	return obj.Items, nil
}

func decodeXmlFile(filePath string, v interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return xml.NewDecoder(file).Decode(v)
}

type securityInfoDirectory struct {
	items map[string]core.SecurityInfo
}

func NewSecurityInfoDirectory(securityInfoStorage core.SecurityInfoStorage) *securityInfoDirectory {
	var items = make(map[string]core.SecurityInfo)
	var ss, err = securityInfoStorage.ReadAll()
	if err == nil {
		for _, s := range ss {
			items[s.SecurityCode] = s
		}
	}
	return &securityInfoDirectory{items}
}

func (srv *securityInfoDirectory) Read(securityCode string) (core.SecurityInfo, bool) {
	var item, found = srv.items[securityCode]
	return item, found
}
