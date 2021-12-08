package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	ldb "github.com/reviashko/parser/localdb"
	"github.com/reviashko/parser/model"
	mod "github.com/reviashko/parser/model"
)

//SupplierHandler struct
type SupplierHandler struct {
	LoaclDB     ldb.LocalDB
	SuppService mod.Supplier
}

//SetOrgRequest func
func (c *SupplierHandler) SetOrgRequest(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintf(w, "unmarshal ReadAll error")
		return
	}

	var suppItem model.SupplierItem
	err = json.Unmarshal([]byte(string(body)), &suppItem)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintf(w, `{"accepted":false, "reason":"SaveError"}`)
		return
	}

	err = c.SuppService.SetOrgRequest(suppItem)
	if err != nil {
		log.Println("GetOrgData error")
		fmt.Fprintf(w, `{"accepted":false, "reason":"%s"}`, err.Error())
		return
	}

	fmt.Fprintf(w, `{"accepted":true}`)
	//fmt.Fprintf(w, `{"kpp":"500301001", "ogrn":"1067746062449", "addr":"Московская обл, Ленинский р-н, деревня Мильково, влд 1", "city":"", "chief":"Бакальчук Татьяна Владимировна", "org":"ВАЙЛДБЕРРИЗ"}`)

}

//GetOffer func
func (c *SupplierHandler) GetOffer(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	//vars := mux.Vars(r)
	//offerID := vars["offer_id"]

	obj, err := json.Marshal(c.LoaclDB.GetData(5))
	if err != nil {
		log.Println("GetOffer error")
		return
	}

	fmt.Fprintf(w, "%s", string(obj))

}

//GetOrgData func
func (c *SupplierHandler) GetOrgData(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)

	inn, err := strconv.ParseInt(vars["inn"], 10, 64)
	if err != nil {
		fmt.Printf("GetOrgData: wrong inn format [%s]", err.Error())
		return
	}

	orgdata, err := c.SuppService.GetDaData(inn)
	if err != nil {
		fmt.Printf("GetOrgData: GetData error [%s]", err.Error())
		return
	}

	fmt.Fprintf(w, "%s", orgdata)
	//fmt.Fprintf(w, `{"kpp":"500301001", "ogrn":"1067746062449", "addr":"Московская обл, Ленинский р-н, деревня Мильково, влд 1", "city":"", "chief":"Бакальчук Татьяна Владимировна", "org":"ВАЙЛДБЕРРИЗ"}`)

}
