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

const KeyPublic = "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAstKpYuR1uxUyr3xwjWsd\n3lxMY79ME1ZQlALyHsRao3lP4F419bD9Q4Rvv8S3Irb6cpFLZgFn+U1/6alNuMjk\nwc4gQD01Eg23OJXtzaXmxg+b4TrLZJadOXHEGczOb4qIofTM4SP+tSNcLSxJJDdJ\nYyFOiZ6UqZi7dNlM6aGU5GKxGnK4ND2FpuaI5rv3isyCV71BRk9kv8wzKYTy3dO+\noTl0EZ0vHDWA8rXqL17WKsrHdx0u0dLzOSCUv4JEezbnRxDgM/ZvszuOFq8nOqBv\nE2viM1u9Tvq6xPmpMV1tRfG3Y7uK6rZhqorGVHwwbftovAdU/lFqduif4NNhJ3Vq\n8/d6L6CPZH/aojUDAlXGketoiW3aOVZF3ESDKRUHWc5s9PJ//QQSkVWipmeYHCex\nvogTsqBswzHhP1yaXaNA/xJNWUuCtq8r82C40RFkjQC5eAUHUvqlHmBdLk2VN+Zm\nSDuoqMOke2X+6SwahB6DxMke+9grzxhMYgebgLwij6SRoBnG7Kxgh1ozGguRIu+1\nQpFjkyJ3fwf+tlD9qcSaWleXwDQ3s52Hu31S+BDEL5/R1KsGBhQJiSEfkmqhIOfE\nw+rRjoO4BNhJYiipcDuBobZb6OiLXrARWVDSjRiavHACCweLQ5Pj48NfubR41exF\nHtz3bmTMDHrBMXdSJqv974kCAwEAAQ==\n-----END PUBLIC KEY-----"
const KeyPrivate = "-----BEGIN RSA PRIVATE KEY-----\n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\n-----END RSA PRIVATE KEY-----"
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

func generateJWTToken(uid string) (string, error) {
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

func transferFromUserToUser(transferRequest TransferRequest) (bool, error) {
	accessToken, err := generateJWTToken(transferRequest.FromUid)

	if err != nil {
		return false, err
	}

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

func main() {
	// LU HONG HAI 0111 222 333 => uid 3
	// Alex Walker 0179 999 999 => uid 177780178

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.POST("/transfer", func(c *gin.Context) {
		var transferRequest TransferRequest

		err := c.BindJSON(&transferRequest)

		if err != nil {
			panic(err)
		}

		log.Println(transferRequest.FromUid)
		log.Println(transferRequest.ToUid)

		result, err := transferFromUserToUser(TransferRequest{
			FromUid: transferRequest.FromUid,
			ToUid:   transferRequest.ToUid,
			Pin:     transferRequest.Pin,
			Amount:  transferRequest.Amount,
			Message: transferRequest.Message,
		})

		if err != nil || !result {
			c.JSON(http.StatusServiceUnavailable, Meta{Code: http.StatusServiceUnavailable, Message: "Something went wrong ahuhu"})
			return
		}

		c.JSON(http.StatusOK, Meta{Code: 200, Message: "DONE, BABE"})
	})

	_ = router.Run(":" + port)
}
