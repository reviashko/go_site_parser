package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/reviashko/parser/exthandlers"

	"github.com/tkanos/gonfig"

	cap "github.com/afocus/captcha"
	"github.com/reviashko/logopassapi/auth"
	"github.com/reviashko/logopassapi/controller"
	ctrl "github.com/reviashko/logopassapi/controller"
	"github.com/reviashko/logopassapi/models"
	"github.com/reviashko/logopassapi/utils"
	parser "github.com/reviashko/parser/getter"
	"github.com/reviashko/parser/handlers"
	ldb "github.com/reviashko/parser/localdb"
	"github.com/reviashko/parser/model"
	mod "github.com/reviashko/parser/model"
)

func main() {

	connectionData := models.ConnectionData{}
	if gonfig.GetConf("config/db.json", &connectionData) != nil {
		log.Panic("load db confg error")
	}

	frontSettings := auth.FrontSettings{}
	if gonfig.GetConf("config/front.json", &frontSettings) != nil {
		log.Panic("load front confg error")
	}

	daDataSettings := model.DaDataKey{}
	if gonfig.GetConf("config/dadata.json", &daDataSettings) != nil {
		log.Panic("load dadata confg error")
	}

	smtpData := utils.SMTPData{}
	if gonfig.GetConf("config/smtp.json", &smtpData) != nil {
		log.Panic("load smtp confg error")
	}

	cryptoData := auth.CryptoData{}
	if gonfig.GetConf("config/crypto.json", &cryptoData) != nil {
		log.Panic("load crypto confg error")
	}

	db, err := models.InitDB(connectionData)
	if err != nil {
		log.Panic(err.Error())
	}

	captcha := handlers.Captcha{Instance: cap.New(), CryptoService: cryptoData, FrontSettings: frontSettings}
	if err := captcha.Init(); err != nil {
		log.Panic("captcha init error")
	}

	var dataObtainer parser.Parser
	dataObtainer.Init()

	var suppService mod.Supplier
	err = suppService.Init(daDataSettings.APIKey, daDataSettings.SecretKey, db)
	if err != nil {
		log.Panic(err.Error())
	}

	var localDB ldb.LocalDB
	localDB.Init("output.json")
	//dataObtainer.Parse("output2.json")

	cntrl := ctrl.NewController(db, cryptoData, smtpData, frontSettings)
	router := cntrl.NewRouter()

	//nonauth
	suppHandler := handlers.SupplierHandler{LoaclDB: localDB, SuppService: suppService}
	router.HandleFunc("/offer/{offer_id}", suppHandler.GetOffer).Methods("GET")

	router.HandleFunc("/orgdata/{inn}", suppHandler.GetOrgData).Methods("GET")
	router.HandleFunc("/orgdata", suppHandler.SetOrgRequest).Methods("POST")

	router.HandleFunc("/captcha/{data}", captcha.Get).Methods("GET")
	router.HandleFunc("/captchaurl", captcha.GetURL).Methods("GET")

	//auth
	extSuppHandler := exthandlers.SupplierHandler{SuppService: suppService}
	authSuppCall := controller.ExternalCall{Cntrl: cntrl, FrontSettings: frontSettings, ExternalLogic: &extSuppHandler}
	router.HandleFunc("/supp/{inn}", authSuppCall.CheckTokenAndDoFunc).Methods("GET", "OPTIONS", "POST", "PUT")

	//7552e535
	fmt.Println("Started...")
	log.Fatal(http.ListenAndServe(":8080", router))
	//cgi.Serve(nil)
}
