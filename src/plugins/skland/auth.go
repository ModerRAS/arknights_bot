package skland

import (
	bot "arknights_bot/config"
	"fmt"
	"github.com/starudream/go-lib/core/v2/gh"
	"github.com/tidwall/gjson"
	"log"
)

type GrantAppData struct {
	Uid  string `json:"uid"`
	Code string `json:"code"`
}

type GenCredByCodeData struct {
	UserId string `json:"userId"`
	Cred   string `json:"cred"`
	Token  string `json:"token"`
}

type AuthRefreshData struct {
	Token string `json:"token"`
}

type ListPlayerData struct {
	List []*PlayersByApp `json:"list"`
}

type PlayersByApp struct {
	AppCode     string    `json:"appCode"`
	AppName     string    `json:"appName"`
	DefaultUid  string    `json:"defaultUid"`
	BindingList []*Player `json:"bindingList"`
}

type Player struct {
	Uid             string `json:"uid"`
	ChannelName     string `json:"channelName"`
	ChannelMasterId string `json:"channelMasterId"`
	NickName        string `json:"nickName"`
	IsOfficial      bool   `json:"isOfficial"`
	IsDefault       bool   `json:"isDefault"`
	IsDelete        bool   `json:"isDelete"`
}

type GenTokenByUidData struct {
	Token string `json:"token"`
}

type User struct {
	HgId string `json:"hgId"`
}

// Login 使用token登录
func Login(token string) (Account, error) {
	account := Account{}

	if token == "" {
		return account, fmt.Errorf("token is empty")
	}
	account.Hypergryph.Token = token

	res, err := grantApp(token, "4ca99fa6b56cc2ba")
	if err != nil {
		return account, fmt.Errorf("grant app error: %w", err)
	}
	account.Hypergryph.Code = res.Code

	res1, err := authLoginByCode(res.Code)
	if err != nil {
		return account, fmt.Errorf("auth login by code error: %w", err)
	}
	u, _ := CheckToken(token)
	account.UserId = u.HgId
	account.Skland.Cred = res1.Cred
	account.Skland.Token = res1.Token
	return account, nil
}

// 获取 OAuth2 授权代码
func grantApp(token string, code string) (*GrantAppData, error) {
	req := HR().SetBody(gh.M{"type": 0, "token": token, "appCode": code})
	return HypergryphRequest[*GrantAppData](req, "POST", "/user/oauth2/v2/grant")
}

// 获取Cred
func authLoginByCode(code string) (*GenCredByCodeData, error) {
	req := SKR().SetHeader("did", did).SetBody(gh.M{"kind": 1, "code": code})
	return SklandRequest[*GenCredByCodeData](req, "POST", "/web/v1/user/auth/generate_cred_by_code")
}

// RefreshToken 刷新 token
func RefreshToken(account Account) (Account, error) {
	res, err := authRefresh(account.Skland.Cred)
	if err != nil {
		return account, fmt.Errorf("auth refresh error: %w", err)
	}
	account.Skland.Token = res.Token
	// 检查cred是否有效
	_, err = listPlayer(account.Skland)
	if err != nil {
		log.Println("cred失效，尝试重新登录。")
		_, err = CheckToken(account.Hypergryph.Token)
		if err != nil {
			return account, err
		}
		account, err = Login(account.Hypergryph.Token)
		if err != nil {
			return account, err
		}
		// 更新token
		bot.DBEngine.Exec("update user_account set hypergryph_token = ?, skland_token = ?, skland_cred = ? where skland_id = ?", account.Hypergryph.Token, account.Skland.Token, account.Skland.Cred, account.UserId)
	}
	return account, nil
}

// CheckToken 检查token有效性
func CheckToken(token string) (*User, error) {
	req := HR().SetQueryParam("token", token)
	user, err := HypergryphRequest[*User](req, "GET", "/user/info/v1/basic")
	if err != nil {
		return nil, fmt.Errorf("token已失效请重新登录！")
	}
	return user, err
}

// 刷新 auth
func authRefresh(cred string) (*AuthRefreshData, error) {
	req := SKR().SetHeader("cred", cred)
	return SklandRequest[*AuthRefreshData](req, "GET", "/api/v1/auth/refresh")
}

// 获取绑定用户列表
func listPlayer(skland AccountSkland) (*ListPlayerData, error) {
	return SklandRequest[*ListPlayerData](SKR(), "GET", "/api/v1/game/player/binding", skland)
}

// ArknightsPlayers 获取明日方舟绑定角色
func ArknightsPlayers(skland AccountSkland) ([]*Player, error) {
	var players []*Player
	playerList, err := listPlayer(skland)
	if err != nil {
		return players, err
	}
	for _, player := range playerList.List {
		if player.AppCode == "arknights" {
			players = player.BindingList
		}
	}
	return players, nil
}

// LoginHypergryph  官网登录
func LoginHypergryph(token, uid string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	reqGrantApp := HR().SetBody(gh.M{"type": 1, "token": token, "appCode": "be36d44aa36bfb5b"})
	grantAppToken, err := HypergryphASRequest(reqGrantApp, "POST", "/user/oauth2/v2/grant")
	if err != nil {
		return "", fmt.Errorf("grant app error: %w", err)
	}

	reqU8Token := HR().SetBody(gh.M{"token": gjson.Parse(grantAppToken).Get("data.token").String(), "uid": uid})
	res, err := reqU8Token.Execute("POST", "https://binding-api-account-prod.hypergryph.com/account/binding/v1/u8_token_by_uid")
	if err != nil {
		return "", fmt.Errorf("get u8token error: %w", err)
	}
	u8Token := gjson.Parse(string(res.Body())).Get("data.token").String()

	reqLogin := HR().SetBody(gh.M{"share_by": "", "share_type": "", "source_from": "", "token": u8Token})
	resLogin, err := HypergryphAKRequest(reqLogin, "POST", "/user/api/role/login")
	if err != nil || resLogin == "" {
		return "", fmt.Errorf("登录失败：%w", err)
	}

	return u8Token, nil
}
