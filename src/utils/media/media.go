package media

import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"golang.org/x/image/webp"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

var imgClient = &http.Client{Timeout: 15 * time.Second}

func GetImg(url string) []byte {
	var pic []byte
	times := 0
	for times < 3 {
		resp, err := imgClient.Get(url)
		if err != nil {
			log.Println("获取图片失败", err)
			times++
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("获取图片失败，状态码：%d，URL：%s", resp.StatusCode, url)
			resp.Body.Close()
			times++
			continue
		}
		pic, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Println("读取图片失败", err)
			times++
			continue
		}
		break
	}
	return pic
}

func ImgConvert(url string) []byte {
	pic, err := http.Get(url)
	if err != nil {
		log.Println("获取图片失败", err)
		return nil
	}
	m, err := webp.Decode(pic.Body)
	if err != nil {
		log.Println("解析图片失败", err)
		return nil
	}
	bounds := m.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	newRgba := image.NewRGBA(bounds)
	f := true
	go overtime(&f)
o:
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			if !f {
				log.Println("图片转换超时")
				break o
			}
			colorRgb := m.At(i, j)
			r, g, b, a := colorRgb.RGBA()
			r_uint8 := uint8(r >> 8)
			g_uint8 := uint8(g >> 8)
			b_uint8 := uint8(b >> 8)
			a_uint8 := uint8(a >> 8)

			if a_uint8 > 23 {
				r_uint8 = 255
				g_uint8 = 255
				b_uint8 = 255
			}
			a_uint8 = 255
			newRgba.SetRGBA(i, j, color.RGBA{R: r_uint8, G: g_uint8, B: b_uint8, A: a_uint8})
		}
	}
	if !f {
		return nil
	}
	buf := new(bytes.Buffer)
	png.Encode(buf, newRgba)
	return buf.Bytes()
}

// CutImg 图片裁剪
func CutImg(url string) []byte {
	pic, err := http.Get(url)
	if err != nil {
		log.Println("获取图片失败", err)
		return nil
	}

	var subImage image.Image
	if pic.Header.Get("Content-Type") == "image/webp" {
		m, err := webp.Decode(pic.Body)
		if err != nil {
			log.Println("解析图片失败", err)
			return nil
		}
		rgba := m.(*image.NYCbCrA)
		subImage = rgba.SubImage(image.Rect(0, m.Bounds().Dy(), m.Bounds().Dx(), int(float64(m.Bounds().Dy())/1.5))).(*image.NYCbCrA)
	} else {
		m, _, err := image.Decode(pic.Body)
		if err != nil {
			log.Println("解析图片失败", err)
			return nil
		}
		rgba := m.(*image.NRGBA)
		subImage = rgba.SubImage(image.Rect(0, m.Bounds().Dy(), m.Bounds().Dx(), int(float64(m.Bounds().Dy())/1.5))).(*image.NRGBA)
	}
	buf := new(bytes.Buffer)
	png.Encode(buf, subImage)
	return buf.Bytes()
}

func overtime(f *bool) {
	time.Sleep(time.Second * 10)
	*f = false
}

// OCR OCR识别
func OCR(file io.Reader, lang, engine, sep string) ([]string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.png")
	if err != nil {
		log.Println("创建文件失败")
		return nil, err
	}
	io.Copy(part, file)
	writer.WriteField("language", lang)
	writer.WriteField("FileType", ".Auto")
	writer.WriteField("OCREngine", engine)
	writer.Close()
	request, err := http.NewRequest("POST", "https://api.ocr.space/parse/image", body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Add("Apikey", "helloworld")
	client := http.DefaultClient
	client.Timeout = time.Second * 10
	resp, err := client.Do(request)
	if err != nil {
		log.Println("ocr失败")
		return nil, err
	}
	read, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	result := gjson.ParseBytes(read)
	log.Println("识别结果：", result.String())
	return strings.Split(result.Get("ParsedResults.0.ParsedText").String(), sep), nil
}

// CreateTelegraphPage 创建telegraph页面
func CreateTelegraphPage(content, title string) string {
	api := viper.GetString("api.telegraph")
	request, _ := http.NewRequest("GET", api, nil)
	params := request.URL.Query()
	params.Add("access_token", viper.GetString("telegraph.token"))
	params.Add("title", title)
	params.Add("content", content)
	request.URL.RawQuery = params.Encode()
	response, _ := http.DefaultClient.Do(request)
	readAll, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return ""
	}
	jsonStr := string(readAll)
	log.Println(jsonStr)
	url := gjson.Get(jsonStr, "result.url").String()
	return url
}
