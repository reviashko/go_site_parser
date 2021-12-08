package exthandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/reviashko/logopassapi/auth"
	"github.com/reviashko/logopassapi/utils"
	"github.com/reviashko/parser/model"
	mod "github.com/reviashko/parser/model"
)

//SupplierHandler struct
type SupplierHandler struct {
	SuppService mod.Supplier
}

//GetResult func
func (c *SupplierHandler) GetResult(request *http.Request, token auth.Token) (string, error) {
	retval := ""

	vars := mux.Vars(request)
	tmpINN, err := strconv.ParseInt(vars["inn"], 10, 64)
	if err != nil {
		return "", err
	}

	if request.Method == "GET" {

		if tmpINN > 0 {
			suppItem, err := c.SuppService.GetSupplier(tmpINN)
			if err != nil {
				log.Println(err.Error())
				return "", err
			}

			obj, err := json.Marshal(suppItem)
			if err != nil {
				log.Println(err.Error())
				return "", err
			}

			retval = string(obj)

		} else {
			suppArray := c.SuppService.GetSupplierRequests()
			if err != nil {
				log.Println(err.Error())
				return "", err
			}

			obj, err := json.Marshal(suppArray)
			if err != nil {
				log.Println(err.Error())
				return "", err
			}

			retval = string(obj)
		}

	}

	if request.Method == "PUT" {

		var suppRequest model.SupplierRequest
		err := utils.ConvertBody2JSON(request.Body, &suppRequest)
		if err != nil {
			log.Println(err.Error())
			return "", err
		}

		err = c.SuppService.AcceptSuppRequest(suppRequest)
		if err != nil {
			log.Println(err.Error())
			return "", err
		}

		retval = `{"ok":true}`
	}

	return retval, nil
}
