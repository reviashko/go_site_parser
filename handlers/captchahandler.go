package handlers

import (
	"fmt"
	"image/color"
	"image/png"
	"math/rand"
	"net/http"
	"time"

	"github.com/afocus/captcha"
	"github.com/gorilla/mux"

	"github.com/reviashko/logopassapi/auth"
	"github.com/reviashko/logopassapi/utils"
)

//Captcha struct
type Captcha struct {
	Instance      *captcha.Captcha
	CryptoService auth.CryptoData
	FrontSettings auth.FrontSettings
}

var fontKinds = [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}

//GetValue func
func (c *Captcha) GetValue(size int) []byte {

	rand.Seed(time.Now().UnixNano())
	ikind, retval := 0, make([]byte, size)

	for i := 0; i < size; i++ {
		ikind = rand.Intn(3)
		scope, base := fontKinds[ikind][0], fontKinds[ikind][1]
		retval[i] = uint8(base + rand.Intn(scope))
	}
	return retval
}

//Init func
func (c *Captcha) Init() error {
	if err := c.Instance.SetFont("fonts/comic.ttf"); err != nil {
		return err
	}

	c.Instance.SetSize(128, 64)
	c.Instance.SetDisturbance(captcha.MEDIUM)
	c.Instance.SetFrontColor(color.RGBA{255, 255, 255, 255})
	c.Instance.SetBkgColor(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 153, 0, 255})

	return nil
}

//GetURL func
func (c *Captcha) GetURL(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", c.FrontSettings.Host)

	captchaValue := string(c.GetValue(c.CryptoService.CaptchaSize))

	fmt.Println(captchaValue)

	val, err := c.CryptoService.EncryptTextAES256Base64(fmt.Sprintf(`%s#%d`, captchaValue, time.Now().Unix()+c.CryptoService.CaptchaTTL))
	if err != nil {
		fmt.Fprintf(w, "%s", utils.GetJSONAnswer("",
			false,
			"captcha GetURL err "+err.Error(),
			""))
	}

	fmt.Fprintf(w, "%s", utils.GetJSONAnswer("",
		true,
		"",
		`"`+val+`"`))
}

//Get func
func (c *Captcha) Get(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	encryptedCaptchaValue := vars["data"]

	decryptedCaptchaValue, err := c.CryptoService.DecryptTextAES256(encryptedCaptchaValue)
	if err != nil {
		decryptedCaptchaValue = "xxx"
	}

	var captcha auth.CAPTCHAData
	ttl, captchaValue, err := captcha.Parse(decryptedCaptchaValue)
	if err != nil || ttl-time.Now().Unix() < 0 {
		captchaValue = string(c.GetValue(c.CryptoService.CaptchaSize))
	}

	//img, captchaWord := c.Instance.Create(6, captcha.ALL)
	img := c.Instance.CreateCustom(captchaValue)
	png.Encode(w, img)
}
