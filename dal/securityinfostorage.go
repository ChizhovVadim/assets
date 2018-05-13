package dal

import (
	"encoding/xml"
	"os"

	"github.com/ChizhovVadim/assets/core"
)

type securityInfoStorage struct {
	items map[string]core.SecurityInfo
}

func NewSecurityInfoStorage(path string) *securityInfoStorage {
	var items, _ = load(path)
	if items == nil {
		items = make(map[string]core.SecurityInfo)
	}
	return &securityInfoStorage{items}
}

func (srv *securityInfoStorage) Read(securityCode string) (core.SecurityInfo, bool) {
	var item, found = srv.items[securityCode]
	return item, found
}

func load(path string) (map[string]core.SecurityInfo, error) {
	var obj = struct {
		Items []core.SecurityInfo `xml:"SecurityInfo"`
	}{}
	var err = decodeXmlFile(path, &obj)
	if err != nil {
		return nil, err
	}
	var result = make(map[string]core.SecurityInfo)
	for _, item := range obj.Items {
		result[item.SecurityCode] = item
	}
	return result, nil
}

func decodeXmlFile(filePath string, v interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return xml.NewDecoder(file).Decode(v)
}
