package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/heroku/go-getting-started/token"
	"gopkg.in/resty.v1"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

const BaseUrl = "https://api-uat.vinid.net:6443/"

type TransferRequest struct {
	FromUid string `json:"from_uid"`
	ToUid   string `json:"to_uid"`
	Pin     string `json:"pin"`
	Amount  int    `json:"amount"`
	Message string `json:"message"`
}

type TransferResponse struct {
	Meta
}

type Meta struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type FundTransferRequest struct {
	BeneficiaryUserId string `json:"beneficiary_user_id"`
	Currency          string `json:"currency"`
	Amount            int    `json:"amount"`
	Description       string `json:"description"`
}

type FundTransferData struct {
	TransactionId          string `json:"transaction_id"`
	TransactionReferenceId string `json:"transaction_reference_id"`
	VerificationMethods    int    `json:"verification_methods"`
}

type FundTransferResponse struct {
	Data FundTransferData `json:"data"`
	Meta Meta             `json:"meta"`
}

type FundTransferConfirmRequest struct {
	TransactionId          string `json:"transaction_id"`
	TransactionReferenceId string `json:"transaction_reference_id"`
	Otp                    string `json:"otp"`
	Pin                    string `json:"pin"`
	Description            string `json:"description"`
}

type FundTransferConfirmData struct {
	TransactionId         string `json:"transaction_id"`
	TransactionStatus     string `json:"transaction_status"`
	TransactionFinishTime int64  `json:"transaction_finish_time"`
}

type FundTransferConfirmResponse struct {
	Data FundTransferConfirmData `json:"data"`
	Meta Meta                    `json:"meta"`
}

/**
Make fund transfer with given user_id and amount
*/
func fundTransfer(accessToken string, request FundTransferRequest) (FundTransferData, error) {
	response, err := resty.R().SetResult(FundTransferResponse{}).
		SetHeader("Authorization", accessToken).
		SetHeader("X-Device-ID", "6CfTMX1FDBrpCdoeDXlDzsxxx").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "vinid.uat/12.0-uat Dalvik/2.1.0 (Linux; U; Android 9; Android SDK built for x86 Build/PSR1.180720.093)").
		SetBody(request).
		Post(BaseUrl + "wallet/v1/fundtransfer/action")

	if err != nil {
		return FundTransferData{}, err
	}

	//println(&response)
	log.Println(string(response.Body()))

	fundTransferResponse := response.Result().(*FundTransferResponse)
	fundTransferData := fundTransferResponse.Data
	return fundTransferData, nil
}

/**
Make fun transfer confirm with given transaction_id and OTP
*/
func fundTransferConfirm(accessToken string, request FundTransferConfirmRequest) (FundTransferConfirmData, error) {
	response, err := resty.R().SetResult(FundTransferConfirmResponse{}).
		SetHeader("Authorization", accessToken).
		SetHeader("X-Device-ID", "6CfTMX1FDBrpCdoeDXlDzs").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "vinid.uat/12.0-uat Dalvik/2.1.0 (Linux; U; Android 9; Android SDK built for x86 Build/PSR1.180720.093)").
		SetBody(request).
		Post(BaseUrl + "wallet/v1/fundtransfer/confirm")

	if err != nil {
		return FundTransferConfirmData{}, err
	}
	fundTransferConfirmResponse := response.Result().(*FundTransferConfirmResponse)
	fundTransferConfirmData := fundTransferConfirmResponse.Data
	return fundTransferConfirmData, nil
}

func generateJWTToken(uid string, KeyPublic string, KeyPrivate string) (string, error) {
	token, err := token.NewJWTHelper(KeyPublic, KeyPrivate)

	if err != nil {
		return "", err
	}

	claim := jwt.StandardClaims{
		Audience:  "",
		ExpiresAt: time.Now().Unix() + 60*60*24*30,
		Id:        "",
		IssuedAt:  time.Now().Unix(),
		Issuer:    "VinID",
		NotBefore: 0,
		Subject:   uid,
	}

	return token.Generate(claim)
}

func transferFromUserToUser(transferRequest TransferRequest, accessToken string) (bool, error) {
	//accessToken, err := generateJWTToken(transferRequest.FromUid)

	//if err != nil {
	//	return false, err
	//}

	fundTransferData, err := fundTransfer(accessToken, FundTransferRequest{
		BeneficiaryUserId: transferRequest.ToUid,
		Currency:          "VND",
		Amount:            transferRequest.Amount,
		Description:       transferRequest.Message,
	})

	if err != nil {
		return false, err
	}

	_, err = fundTransferConfirm(accessToken, FundTransferConfirmRequest{
		TransactionId:          fundTransferData.TransactionId,
		TransactionReferenceId: fundTransferData.TransactionReferenceId,
		Otp:                    "",
		Pin:                    transferRequest.Pin,
		Description:            transferRequest.Message,
	})

	if err != nil {
		return false, err
	}

	return true, err
}

func TokenMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		givenToken := c.Request.Header.Get("Authorization")
		log.Println(givenToken)
		//_, err := base64.StdEncoding.DecodeString(givenToken)
		//if err != nil {
		//	_ = c.AbortWithError(http.StatusUnauthorized, err)
		//}
		//log.Println(string(authInfo))
	}
}

var accounts = gin.Accounts{
	"mentor": "123456",
}

func main() {
	// LU HONG HAI 0111 222 333 => uid 3
	// Alex Walker 0179 999 999 => uid 177780178

	port := os.Getenv("PORT")
	KeyPublic := os.Getenv("JWT_PUBLIC_KEY")
	KeyPrivate := os.Getenv("JWT_PRIVATE_KEY")

	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	authorized := router.Group("/", gin.BasicAuth(accounts))

	authorized.Use(TokenMiddleWare())

	authorized.POST("/transfer", func(c *gin.Context) {
		var transferRequest TransferRequest

		err := c.BindJSON(&transferRequest)

		if err != nil {
			panic(err)
		}

		log.Println(transferRequest.FromUid)
		log.Println(transferRequest.ToUid)

		accessToken, err := generateJWTToken(transferRequest.FromUid, KeyPublic, KeyPrivate)

		if err != nil {
			c.Abort()
		}

		result, err := transferFromUserToUser(TransferRequest{
			FromUid: transferRequest.FromUid,
			ToUid:   transferRequest.ToUid,
			Pin:     transferRequest.Pin,
			Amount:  transferRequest.Amount,
			Message: transferRequest.Message,
		}, accessToken)

		if err != nil || !result {
			c.JSON(http.StatusServiceUnavailable, Meta{Code: http.StatusServiceUnavailable, Message: "Something went wrong ahuhu"})
			return
		}

		c.JSON(http.StatusOK, Meta{Code: 200, Message: "DONE, BABE"})
	})

	authorized.POST("/token", func(c *gin.Context) {
		var accessTokenRequest UserInfoRequest
		err := c.BindJSON(&accessTokenRequest)
		if err != nil {
			c.Abort()
		}

		accessToken, err := generateJWTToken(accessTokenRequest.Uid, KeyPublic, KeyPrivate)

		if err != nil {
			c.Abort()
		}

		c.JSON(http.StatusOK, AccessTokenResponse{AccessToken: accessToken})
	})

	authorized.POST("/wallet/info", func(c *gin.Context) {
		var userInfoRequest UserInfoRequest
		err := c.BindJSON(&userInfoRequest)
		if err != nil {
			c.Abort()
		}

		accessToken, err := generateJWTToken(userInfoRequest.Uid, KeyPublic, KeyPrivate)

		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Can not fetch wallet info"})
		}

		walletAccount, err := getWalletInfo(accessToken)

		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Can not fetch wallet info"})
		}

		c.JSON(http.StatusOK, walletAccount)
	})

	_ = router.Run(":" + port)
}

// temp info

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type UserInfoRequest struct {
	Uid string `json:"uid"`
}

type WalletAccount struct {
	AccountId string `json:"account_id"`
	Currency  string `json:"currency"`
	Balance   int    `json:"balance"`
	CreatedAt int64  `json:"created_at"`
	Status    string `json:"status"`
}

type WalletData struct {
	Accounts []WalletAccount `json:"accounts"`
}

type WalletInfoResponse struct {
	Data WalletData `json:"data"`
	Meta Meta `json:"meta"`
}

func getWalletInfo(accessToken string) (WalletAccount, error) {
	response, err := resty.R().SetResult(WalletInfoResponse{}).
		SetHeader("Authorization", accessToken).
		SetHeader("X-Device-ID", "6CfTMX1FDBrpCdoeDXlDzs").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "vinid.uat/12.0-uat Dalvik/2.1.0 (Linux; U; Android 9; Android SDK built for x86 Build/PSR1.180720.093)").
		Get(BaseUrl + "wallet/v1/wallets")

	if err != nil {
		return WalletAccount{}, err
	}

	//log.Println(string(response.Body()))

	walletInfoResponse := response.Result().(*WalletInfoResponse)
	walletAccount := walletInfoResponse.Data.Accounts[0]

	return walletAccount, nil
}
