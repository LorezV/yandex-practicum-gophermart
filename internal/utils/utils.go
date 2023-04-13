package utils

import (
	"github.com/LorezV/go-diploma.git/internal/config"
	"github.com/LorezV/go-diploma.git/internal/repositories/userrepository"
	"github.com/dgrijalva/jwt-go/v4"
	"time"
)

var DateLayout = "2006-01-02T15:04:05.000Z"

type Claims struct {
	jwt.StandardClaims
	UserID int `json:"user_id"`
}

type ContextKey string

type ContextUser struct {
	User    userrepository.User
	IsValid bool
}

func GenerateUserToken(userID int) (token string, err error) {
	tokenObject := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(time.Hour * 24 * 4)),
			IssuedAt:  jwt.At(time.Now()),
		},
		UserID: userID,
	})

	token, err = tokenObject.SignedString([]byte(config.Config.SecretKey))
	return
}

func ValidLuhn(number int) bool {
	return (number%10+CheckSumLuhn(number/10))%10 == 0
}

func CheckSumLuhn(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}

//func FetchAccrualOrder(number string) (AccrualOrder, error) {
//	resp, err := http.FindUnique(fmt.Sprintf("%s/api/orders/%s", config.Config.AccrualSystemAddress, number))
//
//	resp.
//}
