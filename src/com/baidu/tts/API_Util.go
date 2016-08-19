package tts

import (
	"com/baidu/public"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"org/StevenChen/util"
)

//REST API Url
const API_URL = "http://tsn.baidu.com/text2audio"

type API_Request struct {
	Tex  string `json:"tex"`           //必填；合成文本，使用UTF-8编码，请注意文本长度必须小于1024
	Lan  string `json:"lan"`           //必填；语言选择，填写zh
	Tok  string `json:"tok"`           //必填；开放平台获取到的开发者access_token
	Ctp  int    `json:"ctp"`           //必填；客户端类型选择,web端填写1
	Cuid string `json:"cuid"`          //必填；用户唯一标识，用来区分用户，填写机器 MAC 地址或 IMEI 码，长度为60以内
	Spd  int    `json:"spd,omitempty"` //选填；语速，取值0-9，默认5
	Pit  int    `json:"pit,omitempty"` //选填；语调，取值0-9，默认5
	Vol  int    `json:"vol,omitempty"` //选填；音量，取值0-9，默认5
	Per  int    `json:"per,omitempty"` //选填；发音人选择，取值0-1；默认0女声 1男声
}

var API_ResponseErrEnum = map[int]string{
	500: "不支持输入",
	501: "输入参数不正确",
	502: "token验证失败",
	503: "合成后端错误",
}

type API_Response struct {
	Err_no  int    `json:"err_no"`
	Err_msg string `json:"err_msg"`
	Sn      string `json:"sn"`
	Idx     int    `json:"idx"`
}

type API_Util struct {
	Credentials public.Credentials_Response
	Cuid        string
	api_key     string
	secret_key  string
}

func NewAPI_Util(api_key, secret_key string) API_Util {

	cuid := util.GetCUID()

	res := public.GetCredentials(public.Credentials_Request{
		Client_id: api_key, Client_secret: secret_key})

	var util API_Util
	util.Cuid = cuid
	util.Credentials = res
	util.api_key = api_key
	util.secret_key = secret_key

	return util
}

func (this API_Util) Text2AudioFile(filePath, text string) {

	body := this.Text2AudioBytes(text)

	err := ioutil.WriteFile(filePath, body, 0666)
	if err != nil {
		panic(err.Error())
	}
}

func (this *API_Util) Text2AudioBytes(text string) []byte {

	param := url.Values{}
	param.Set("ctp", "1")
	param.Set("lan", "zh")
	param.Set("tex", text)
	param.Set("cuid", this.Cuid)
	param.Set("tok", this.Credentials.Access_token)

	response, err := http.PostForm(API_URL, param)
	defer response.Body.Close()
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	contentType := response.Header.Get("Content-type")
	if "audio/mp3" == contentType {
		return body
	} else {
		var errMsg API_Response
		err = json.Unmarshal(body, &errMsg)
		if nil != err {
			panic(err.Error())
		} else {
			if 502 == errMsg.Err_no {
				*this = NewAPI_Util(this.api_key, this.secret_key)
				return this.Text2AudioBytes(text)
			} else if errMean, ok := API_ResponseErrEnum[errMsg.Err_no]; ok {
				panic(errMean)
			} else {
				panic("Unknow:API返回的错误编码未定义 err_no:" + strconv.Itoa(errMsg.Err_no))
			}
		}
	}
}