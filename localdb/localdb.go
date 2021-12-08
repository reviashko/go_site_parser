package localdb

import (
	"encoding/json"
	"io/ioutil"

	parser "github.com/reviashko/parser/getter"
)

//LocalDB struct
type LocalDB struct {
	Products map[string]parser.Product
}

//Init func
func (d *LocalDB) Init(filename string) error {

	d.Products = make(map[string]parser.Product)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	tmpData := make([]parser.Product, 0)
	err = json.Unmarshal(data, &tmpData)
	if err != nil {
		return err
	}

	for _, item := range tmpData {
		d.Products[item.Articul] = item
	}

	return nil
}

//GetData func
func (d *LocalDB) GetData(count int) []parser.Product {

	retval := make([]parser.Product, 0)

	rownumver := 0
	for _, item := range d.Products {
		retval = append(retval, item)

		rownumver++
		if rownumver >= count {
			break
		}
	}

	return retval
}
