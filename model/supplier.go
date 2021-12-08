package model

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/reviashko/logopassapi/models"

	b64 "encoding/base64"
	"encoding/json"

	"gopkg.in/webdeskltd/dadata.v2"
)

//SupplierItem struct
type SupplierItem struct {
	City             string   `json:"city"  db:"city"`
	IsEnterprener    bool     `json:"is_entpr"  db:"is_entpr"`
	ContactName      string   `json:"contact_name"  db:"contact_name"`
	ContactPhone     []int64  `json:"contact_phones"  db:"contact_phones"`
	Email            string   `json:"email"  db:"email"`
	Assortiment      []string //list? //ассортимент магазинов
	TradeFacility    string   `json:"shop"` //Торговые площади: ТОЧНЫЙ адрес (город, улица, дом, корпус/строение, название ТЦ, этаж, название магазина), телефон.
	OrgName          string   `json:"org_name"  db:"org_name"`
	Chief            string   `json:"chief"  db:"chief"`
	Accountant       string   `json:"accountant"  db:"accountant"`
	JurAddr          string   `json:"jur_addr"  db:"jur_addr"`
	FactAddr         string   `json:"fact_addr"  db:"fact_addr"`
	PostAddr         string   `json:"post_addr"  db:"post_addr"`
	OGRN             int64    `json:"ogrn"  db:"ogrn"`
	INN              int64    `json:"inn"  db:"inn"`
	KPP              int64    `json:"kpp"  db:"kpp"`
	Bank             string   `json:"bank"  db:"bank"`
	BIK              int64    `json:"bik"  db:"bik"`
	BankAccaunt      string   `json:"bank_accaunt"  db:"bank_accaunt"`
	Accaunt          string   `json:"accaunt"  db:"accaunt"`     //Корр. счет
	TransportCompony string   `json:"transport"  db:"transport"` //list? Транспортная компания
}

//Assortiment      []string //list? //ассортимент магазинов
//TradeFacility    string   `json:"shop"` //Торговые площади: ТОЧНЫЙ адрес (город, улица, дом, корпус/строение, название ТЦ, этаж, название магазина), телефон.
//TransportCompony string   `json:"transport"` //list? Транспортная компания

//Supplier struct
type Supplier struct {
	DaDataService *dadata.DaData
	DaDataCache   map[int64]SupplierItem
	Data          map[int64]SupplierItem
	DB            *models.DB
	SuppRequest   map[int64]SupplierRequest
}

//DaDataKey struct
type DaDataKey struct {
	APIKey    string
	SecretKey string
}

//SupplierRequest struct
type SupplierRequest struct {
	Data     string    `db:"data"`
	Dt       time.Time `json:"dt"  db:"dt"`
	INN      int64     `json:"inn"  db:"inn"`
	Accepted bool      `json:"is_accepted"  db:"is_accepted"`
	OrgName  string    `json:"org_name"  db:"org_name"`
	City     string    `json:"city"  db:"city"`
	Chief    string    `json:"chief"  db:"chief"`
}

//GetSupplierRequest func
func (s *SupplierItem) GetSupplierRequest() SupplierRequest {
	return SupplierRequest{Chief: s.Chief, City: s.City, OrgName: s.OrgName, Accepted: false, INN: s.INN, Dt: time.Now()}
}

//Init func
func (p *Supplier) Init(apiKey string, secretKey string, db *models.DB) error {
	p.DB = db
	p.SuppRequest = make(map[int64]SupplierRequest)
	err := p.UploadSuppliersList()
	if err != nil {
		return err
	}
	p.DaDataService = dadata.NewDaData(apiKey, secretKey)
	p.DaDataCache = make(map[int64]SupplierItem)
	fmt.Println("Suppliers inition is done...")

	return nil
}

//GetSupplier func
func (p *Supplier) GetSupplier(inn int64) (*SupplierItem, error) {
	if item, ok := p.Data[inn]; ok {
		return &item, nil
	}

	return nil, errors.New("supplier not found")
}

//UploadSuppliersList func
func (p *Supplier) UploadSuppliersList() error {

	rows, err := p.DB.Queryx(`select inn, dt, is_accepted, data from supp.request_get()`)
	if err != nil {
		return err
	}
	defer rows.Close()

	p.Data = make(map[int64]SupplierItem)

	for rows.Next() {
		newItem := new(SupplierRequest)
		err = rows.StructScan(&newItem)
		if err != nil {
			return err
		}

		newSuppItem := new(SupplierItem)
		err = json.Unmarshal([]byte(newItem.Data), newSuppItem)
		if err != nil {
			return err
		}

		newItem.OrgName = newSuppItem.OrgName
		newItem.Chief = newSuppItem.Chief
		newItem.City = newSuppItem.City

		p.SuppRequest[newItem.INN] = *newItem
		p.Data[newSuppItem.INN] = *newSuppItem
	}

	return nil
}

//GetSupplierRequests func
func (p *Supplier) GetSupplierRequests() []SupplierRequest {
	retval := make([]SupplierRequest, 0)

	for _, item := range p.SuppRequest {
		retval = append(retval, item)
	}

	return retval
}

//AcceptSuppRequest func
func (p *Supplier) AcceptSuppRequest(suppRequest SupplierRequest) error {

	_, err := p.DB.Queryx("SELECT supp.request_accept($1::bigint, $2)", suppRequest.INN, suppRequest.Accepted)
	if err != nil {
		return err
	}

	if item, ok := p.SuppRequest[suppRequest.INN]; ok {
		item.Accepted = suppRequest.Accepted
		p.SuppRequest[suppRequest.INN] = item
	}

	return nil
}

//SetOrgRequest func
func (p *Supplier) SetOrgRequest(suppItem SupplierItem) error {
	if _, ok := p.Data[suppItem.INN]; ok {
		return errors.New("SupplierExist")
	}

	suppItem.OrgName = b64.StdEncoding.EncodeToString([]byte(suppItem.OrgName))
	suppItem.JurAddr = b64.StdEncoding.EncodeToString([]byte(suppItem.JurAddr))
	suppItem.PostAddr = b64.StdEncoding.EncodeToString([]byte(suppItem.PostAddr))
	suppItem.FactAddr = b64.StdEncoding.EncodeToString([]byte(suppItem.FactAddr))

	obj, err := json.Marshal(suppItem)
	if err != nil {
		return err
	}

	_, err = p.DB.Queryx("SELECT supp.request_save($1::bigint, $2::json)", suppItem.INN, string(obj))
	if err != nil {
		return err
	}

	//update campaigns data
	p.Data[suppItem.INN] = suppItem
	p.SuppRequest[suppItem.INN] = suppItem.GetSupplierRequest()

	return nil
}

//GetDaData func
func (p *Supplier) GetDaData(inn int64) (string, error) {

	supplierExists := false

	if _, ok := p.Data[inn]; ok {
		supplierExists = true
	}

	if item, ok := p.Data[inn]; ok {
		return fmt.Sprintf(`{"kpp":"%d", "ogrn":"%d", "addr":"%s", "city":"%s", "chief":"%s", "org":"%s", "is_entpr":%t, "exists":%t}`, item.KPP, item.OGRN, item.JurAddr, item.City, item.Chief, item.OrgName, item.IsEnterprener, supplierExists), nil
	}

	var retval SupplierItem

	if mapItem, ok := p.DaDataCache[inn]; ok {
		retval = mapItem
	} else {

		orgs, err := p.DaDataService.SuggestParties(dadata.SuggestRequestParams{Query: fmt.Sprintf("%d", inn), Count: 3})
		if nil != err || len(orgs) == 0 {
			return "", err
		}

		if orgs == nil || len(orgs) == 0 {
			return "", nil
		}

		tmpKPP, err := strconv.ParseInt("0"+orgs[0].Data.Kpp, 10, 64)
		if err != nil {
			return "", err
		}

		tmpOGRN, err := strconv.ParseInt("0"+orgs[0].Data.Ogrn, 10, 64)
		if err != nil {
			return "", err
		}

		tmpRetval := new(SupplierItem)
		tmpRetval.INN = inn
		tmpRetval.KPP = tmpKPP
		tmpRetval.OGRN = tmpOGRN
		tmpRetval.JurAddr = b64.StdEncoding.EncodeToString([]byte(orgs[0].Data.Address.Value))
		tmpRetval.City = orgs[0].Data.Address.Data.City
		tmpRetval.Chief = orgs[0].Data.Management.Name
		tmpRetval.OrgName = b64.StdEncoding.EncodeToString([]byte(orgs[0].Data.Name.ShortWithOpf))
		tmpRetval.IsEnterprener = orgs[0].Data.Type == "INDIVIDUAL"

		p.DaDataCache[inn] = *tmpRetval
		retval = *tmpRetval
	}

	//change to unmarshal
	return fmt.Sprintf(`{"kpp":"%d", "ogrn":"%d", "addr":"%s", "city":"%s", "chief":"%s", "org":"%s", "is_entpr":%t, "exists":%t}`, retval.KPP, retval.OGRN, retval.JurAddr, retval.City, retval.Chief, retval.OrgName, retval.IsEnterprener, supplierExists), nil
}
