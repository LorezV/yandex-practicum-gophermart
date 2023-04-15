package handlers

import (
	"errors"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/LorezV/go-diploma.git/internal/services"
	"github.com/LorezV/go-diploma.git/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strconv"
)

type Handler struct {
	services *services.Services
}

func MakeHandler(services *services.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) Register(c echo.Context) error {
	type RequestData struct {
		Login    string `json:"login" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	requestData := new(RequestData)
	if err := c.Bind(requestData); err != nil {
		return err
	}

	if err := validator.New().Struct(requestData); err != nil {
		err := err.(validator.ValidationErrors)[0]
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("The %s is %s", err.Field(), err.Tag()))
	}

	user, err := h.services.User.Create(c.Request().Context(), requestData.Login, requestData.Password)
	if err != nil {
		if errors.Is(err, services.ErrLoginTaken) {
			return echo.NewHTTPError(http.StatusConflict, "This login is already taken")
		}

		log.Error().Err(err).Msg("Create user")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	token, err := h.services.Auth.GenerateToken(user)
	if err != nil {
		log.Error().Err(err).Msg("Login user")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	c.Response().Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return c.NoContent(http.StatusOK)
}

func (h *Handler) Login(c echo.Context) error {
	type RequestData struct {
		Login    string `json:"login" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	requestData := new(RequestData)
	if err := c.Bind(requestData); err != nil {
		return err
	}

	if err := validator.New().Struct(requestData); err != nil {
		err := err.(validator.ValidationErrors)[0]
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("The %s is %s", err.Field(), err.Tag()))
	}

	token, err := h.services.Auth.Login(c.Request().Context(), requestData.Login, requestData.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "These RequestData don't match our records")
		}

		log.Error().Err(err).Msg("Login user")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	c.Response().Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return c.NoContent(http.StatusOK)
}

func (h *Handler) PostOrders(c echo.Context) error {
	numberBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Error().Err(err).Msg("Read body on creating order")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	number := string(numberBytes)
	numberInt, err := strconv.Atoi(number)
	if err != nil {
		log.Error().Err(err).Msg("Converting body bytes to int")
		return echo.NewHTTPError(http.StatusBadRequest, "The number is required to be passed as a plain text")
	}

	if !utils.ValidLuhn(numberInt) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Invalid order numberBytes")
	}

	user := c.Get("user").(*models.User)
	order, err := h.services.Order.FindByNumber(c.Request().Context(), number)
	if err != nil {
		log.Error().Err(err).Msg("FindByID order by numberBytes")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if order != nil {
		if order.UserID == user.ID {
			return echo.NewHTTPError(http.StatusOK, "The order has already been taken")
		} else {
			return echo.NewHTTPError(http.StatusConflict, "The order number is already registered by another user")
		}
	}

	if _, err = h.services.Order.Create(c.Request().Context(), number, user); err != nil {
		log.Error().Err(err).Msg("Create order")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusAccepted, "The order has been accepted")
}

func (h *Handler) GetOrders(c echo.Context) error {
	user := c.Get("user").(*models.User)

	orders, err := h.services.Order.FindAll(c.Request().Context(), user)
	if err != nil {
		log.Error().Err(err).Msg("Getting all user orders")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if len(orders) > 0 {
		return c.JSON(http.StatusOK, utils.MakeOrdersResponse(orders))
	} else {
		return c.NoContent(http.StatusNoContent)
	}
}

func (h *Handler) GetBalance(c echo.Context) error {
	user := c.Get("user").(*models.User)

	type userBalance struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}

	withdrawn, err := h.services.Withdrawal.Sum(c.Request().Context(), user)
	if err != nil {
		log.Error().Err(err).Msg("Withdrawal sum")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, userBalance{
		Current:   user.Balance,
		Withdrawn: withdrawn,
	})
}

func (h *Handler) PostWithdraw(c echo.Context) error {
	type withdrawal struct {
		Order string  `json:"order" validate:"required"`
		Sum   float64 `json:"sum" validate:"required"`
	}

	user := c.Get("user").(*models.User)

	w := new(withdrawal)
	if err := c.Bind(w); err != nil {
		log.Error().Err(err).Msg("Bind withdrawalModel")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := validator.New().Struct(w); err != nil {
		err := err.(validator.ValidationErrors)[0]
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("The %s is %s", err.Field(), err.Tag()))
	}

	number, err := strconv.Atoi(w.Order)
	if err != nil {
		log.Error().Err(err).Msg("Number string to int conversion")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if !utils.ValidLuhn(number) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	withdrawalModel := &models.Withdrawal{
		UserID: user.ID,
		Order:  w.Order,
		Sum:    w.Sum,
	}
	if err = h.services.Withdrawal.Create(c.Request().Context(), withdrawalModel, user); err != nil {
		log.Error().Err(err).Msg("Create withdrawalModel")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetWithdrawals(c echo.Context) error {
	user := c.Get("user").(*models.User)

	withdrawals, err := h.services.Withdrawal.All(c.Request().Context(), user)
	if err != nil {
		log.Error().Err(err).Msg("Getting all user withdrawals")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if len(withdrawals) > 0 {
		return c.JSON(http.StatusOK, utils.MakeWithdrawalsResponse(withdrawals))
	} else {
		return c.NoContent(http.StatusNoContent)
	}
}
