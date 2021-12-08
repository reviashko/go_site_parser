package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

//Product struct
type Product struct {
	Name      string   `json:"name"`
	Articul   string   `json:"articul"`
	URL       string   `json:"url"`
	Sizes     []string `json:"sizes"`
	Spicifies []string `json:"spec"`
	Descr     string   `json:"descr"`
	Photos    []string `json:"photos"`
	Crumbs    []string `json:"crumbs"`
	ExLinks   []string `json:"exlinks"`
}

//Urls struct
type Urls struct {
	URL   string
	Pages int
}

//Parser struct
type Parser struct {
	Products []Product
	Urls     []Urls
}

//Init func
func (p *Parser) Init() {

	p.Products = make([]Product, 0)
	p.Urls = make([]Urls, 0)

	p.Urls = append(p.Urls, Urls{URL: "https://naumi.ru/catalog/detskoe/filter/clear/?PAGEN_2=", Pages: 9})
	p.Urls = append(p.Urls, Urls{URL: "https://naumi.ru/catalog/zhenskoe/filter/clear/?PAGEN_2=", Pages: 41})
	p.Urls = append(p.Urls, Urls{URL: "https://naumi.ru/catalog/filter/aksessuar-is-true/apply/?PAGEN_2=", Pages: 3})
	p.Urls = append(p.Urls, Urls{URL: "https://naumi.ru/specials/filter/clear/apply/?PAGEN_2=", Pages: 57})
}

//Parse func
func (p *Parser) Parse(filename string) {

	for _, item := range p.Urls {
		data, err := getURLProducts(item.URL, item.Pages)
		if err == nil {
			p.Products = append(p.Products, data...)
		} else {
			fmt.Printf("Product Error...[%s]", err.Error())
		}
	}

	productsJSON, err := json.Marshal(p.Products)
	if err == nil {
		err = ioutil.WriteFile(filename, productsJSON, 0644)
		if err == nil {
			fmt.Printf("\nDone...")
		} else {
			fmt.Printf("\nError...")
		}
	} else {
		fmt.Printf("\nJSON Error...")
	}

}

func getURLProducts(url string, pages int) ([]Product, error) {

	retval := make([]Product, 0)

	for i := 1; i <= pages; i++ {

		trycount := 0
		for {
			trycount++

			fmt.Printf("\nUrl: [%s%d] Attemp:[%d] Items:", url, i, trycount)

			time.Sleep(2 * time.Second)

			doc, err := htmlquery.LoadURL(fmt.Sprintf("%s%d", url, i))
			if err != nil {
				return retval, err
			}

			contentBody := htmlquery.FindOne(doc, "//div[@class='row  js--ajax-load-var1__parent__items']")
			itemsList := htmlquery.Find(contentBody, "div")

			itemsObtained := 0

			if itemsList != nil {

				ch := make(chan Product, len(itemsList))
				for _, product := range itemsList {
					go getProductData(product, ch)
				}

				for i := 1; i <= len(itemsList); i++ {
					data := <-ch
					if len(data.URL) > 5 {
						retval = append(retval, data)
						itemsObtained++
					}
				}
			}

			fmt.Printf("[%d]", len(itemsList))

			if itemsObtained > 0 || trycount > 2 {
				break
			}
		}

	}

	return retval, nil
}

func getProductExtraData(product *Product) *Product {

	time.Sleep(500 * time.Microsecond)
	doc, err := htmlquery.LoadURL("https://naumi.ru" + product.URL)
	if err != nil {
		return product
	}

	articul := htmlquery.FindOne(doc, "//div[@class='catalog-element-var1__vendor-code']")
	if articul != nil {
		product.Articul = strings.Trim(spaceStringsBuilder(htmlquery.InnerText(articul)), "Артикул:")
	}

	descr := htmlquery.FindOne(doc, "//div[@class='catalog-element-var1__text-detail']/div[1]/div[1]")
	if descr != nil {
		product.Descr = spaceStringsBuilder(htmlquery.InnerText(descr))
	}

	specTag := htmlquery.FindOne(doc, "//div[@class='catalog-element-var1__specifications']/div[1]/div[1]/div[1]/ul[1]")
	if specTag != nil {
		specList := htmlquery.Find(specTag, "li")
		if specList != nil {

			for _, specItem := range specList {
				product.Spicifies = append(product.Spicifies, spaceStringsBuilder(htmlquery.InnerText(specItem)))
			}
		}
	}

	otherColorsDic := htmlquery.FindOne(doc, "//div[@class='catalog-element-var1__colors__other']/div[2]")
	if otherColorsDic != nil {
		colorList := htmlquery.Find(otherColorsDic, "a")
		if colorList != nil {

			for _, colorItem := range colorList {
				product.ExLinks = append(product.ExLinks, spaceStringsBuilder(htmlquery.SelectAttr(colorItem, "href")))
			}
		}
	}

	crumbsDiv := htmlquery.FindOne(doc, "//div[@class='breadcrumb-var1']")
	if crumbsDiv != nil {
		crumbList := htmlquery.Find(crumbsDiv, "a[@class='breadcrumb-var1__item']")
		if crumbList != nil {

			for _, crumbItem := range crumbList {
				innerText := spaceStringsBuilder(htmlquery.InnerText(crumbItem))
				if innerText != "Главная" && innerText != "Каталог" {
					product.Crumbs = append(product.Crumbs, innerText)
				}
			}
		}
	}

	photoList := htmlquery.Find(doc, "//div[@class='catalog-element-var1__slider-thumbs__slide']")
	if photoList != nil {
		for _, photoItem := range photoList {

			imgParts := strings.Split(htmlquery.SelectAttr(photoItem, "style"), "(")
			if imgParts != nil && len(imgParts) >= 1 {
				product.Photos = append(product.Photos, strings.Trim(imgParts[1], ");"))
			}
		}
	}

	return product
}

func getProductData(product *html.Node, c chan Product) {

	newItem := new(Product)

	rootItemDIV := htmlquery.FindOne(product, "div[1]/div[1]/div[2]")
	nameDiv := htmlquery.FindOne(rootItemDIV, "div[@class='catalog-item-var1__preview__name']")

	href := htmlquery.FindOne(product, "div[1]/div[1]/div[3]/div[1]/a")
	if href != nil {
		sizesDiv := htmlquery.FindOne(product, "div[1]/div[1]/div[3]/div[1]/div[1]")

		sizes := make([]string, 0)

		if sizesDiv != nil {
			sizeList := htmlquery.Find(sizesDiv, "div")

			if sizeList != nil {
				for _, size := range sizeList {
					sizes = append(sizes, spaceStringsBuilder(htmlquery.InnerText(size)))
				}
			}
		}

		newItem.URL = htmlquery.SelectAttr(href, "href")
		newItem.Name = spaceStringsBuilder(htmlquery.InnerText(nameDiv))
		newItem.Sizes = sizes

		c <- *getProductExtraData(newItem)
	}

	c <- *newItem
}

func spaceStringsBuilder(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) || ch == 32 {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func getHTML(url string) ([]byte, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	return html, nil
}
