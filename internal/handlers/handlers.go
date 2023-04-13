package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/config"
	"github.com/LorezV/go-diploma.git/internal/repositories/orderrepository"
	"github.com/LorezV/go-diploma.git/internal/repositories/userrepository"
	"github.com/LorezV/go-diploma.git/internal/repositories/withdrawalrepository"
	"github.com/LorezV/go-diploma.git/internal/utils"
	"github.com/jackc/pgerrcode"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var data struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if data.Login == "" || data.Password == "" {
		http.Error(w, "invalid login or password", http.StatusBadRequest)
		return
	}

	password := sha256.New()
	password.Write([]byte(data.Password))
	password.Write([]byte(config.Config.PasswordSalt))
	data.Password = fmt.Sprintf("%x", password.Sum(nil))

	err = userrepository.Create(r.Context(), data.Login, data.Password, config.Config.PasswordSalt)
	if err != nil {
		if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
			http.Error(w, "this login is already occupied", http.StatusConflict)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	user, err := userrepository.FindUnique(r.Context(), "login", data.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := utils.GenerateUserToken(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var data struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if data.Login == "" || data.Password == "" {
		http.Error(w, "invalid value for login or password", http.StatusBadRequest)
		return
	}

	user, err := userrepository.FindUnique(r.Context(), "login", data.Login)
	if err != nil {
		if err.Error() == "no rows in result set" {
			http.Error(w, "invalid login or password", http.StatusUnauthorized)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	password := sha256.New()
	password.Write([]byte(data.Password))
	password.Write([]byte(config.Config.PasswordSalt))
	data.Password = fmt.Sprintf("%x", password.Sum(nil))

	if data.Password != user.Password {
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateUserToken(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
}

func PostOrders(w http.ResponseWriter, r *http.Request) {
	cUser := r.Context().Value(utils.ContextKey("user")).(utils.ContextUser)
	if !cUser.IsValid {
		http.Error(w, "you unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	number, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if ok := utils.ValidLuhn(number); !ok {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}

	_, err = orderrepository.Create(r.Context(), cUser.User.ID, fmt.Sprintf("%d", number))
	if err != nil {
		if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
			order, err := orderrepository.FindUnique(r.Context(), "number", fmt.Sprintf("%d", number))
			if err != nil {
				http.Error(w, err.Error(), http.StatusGone)
				return
			}

			if order.UserID != cUser.User.ID {
				http.Error(w, "order already registered", http.StatusConflict)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	cUser := r.Context().Value(utils.ContextKey("user")).(utils.ContextUser)
	if !cUser.IsValid {
		http.Error(w, "you unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := orderrepository.FindByUser(r.Context(), cUser.User.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type AccrualResponseData struct {
		Number     string  `json:"number"`
		Status     string  `json:"status"`
		Accrual    float64 `json:"accrual"`
		UploadedAt string  `json:"uploaded_at"`
	}

	if len(orders) <= 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	result := make([]AccrualResponseData, 0)
	for _, order := range orders {
		result = append(result, AccrualResponseData{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    *order.Accrual,
			UploadedAt: order.CreatedAt.Format(time.RFC3339),
		})
	}

	responseBody, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	cUser := r.Context().Value(utils.ContextKey("user")).(utils.ContextUser)
	if !cUser.IsValid {
		http.Error(w, "you unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawn, err := withdrawalrepository.Sum(r.Context(), cUser.User.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseData, err := json.Marshal(struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{Current: cUser.User.Balance, Withdrawn: withdrawn})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseData)
}

func PostWithdraw(w http.ResponseWriter, r *http.Request) {
	cUser := r.Context().Value(utils.ContextKey("user")).(utils.ContextUser)
	if !cUser.IsValid {
		http.Error(w, "you unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var withdrawal struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	err = json.Unmarshal(body, &withdrawal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	number, err := strconv.Atoi(withdrawal.Order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if ok := utils.ValidLuhn(number); !ok {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}

	if cUser.User.Balance < withdrawal.Sum {
		http.Error(w, "not enouth money", http.StatusPaymentRequired)
		return
	}

	err = userrepository.Withdraw(r.Context(), cUser.User, withdrawal.Order, withdrawal.Sum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	cUser := r.Context().Value(utils.ContextKey("user")).(utils.ContextUser)
	if !cUser.IsValid {
		http.Error(w, "you unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := withdrawalrepository.AllByUser(r.Context(), cUser.User.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) <= 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	responseBody, err := json.Marshal(withdrawals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}
