package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/swastiijain24/bank/internals/services"
)

type AccountHandler struct {
	accountService services.AccountService
}

func NewAccountHandler(accountService services.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

func (h *AccountHandler) DiscoverAccounts(c *gin.Context) {
	var DiscoverAccountReq discoverAccountReq
	if err := c.ShouldBindJSON(&DiscoverAccountReq); err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	accounts, err := h.accountService.DiscoverAccounts(c.Request.Context(), DiscoverAccountReq.Phone, DiscoverAccountReq.BankCode)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, gin.H{"accounts": accounts})
}

func (h *AccountHandler) GetBalance(c *gin.Context) {
	var GetBalance getBalanceReq
	if err := c.ShouldBindJSON(&GetBalance); err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	balance, err := h.accountService.GetBalance(c.Request.Context(), GetBalance.AccountId, GetBalance.BankCode, GetBalance.MpinEn)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, gin.H{"balance": balance})
}

func (h *AccountHandler) SetMpin(c *gin.Context) {
	var SetMpin setMpinReq
	if err := c.ShouldBindJSON(&SetMpin); err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	err := h.accountService.SetMpin(c.Request.Context(), SetMpin.AccountId, SetMpin.BankCode, SetMpin.MpinEn)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, gin.H{"response": nil})
}

func (h *AccountHandler) ChangeMpin(c *gin.Context) {
	var ChangeMpin changeMpinReq
	if err := c.ShouldBindJSON(&ChangeMpin); err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	err := h.accountService.ChangeMpin(c.Request.Context(), ChangeMpin.AccountId, ChangeMpin.BankCode, ChangeMpin.OldMpinEn, ChangeMpin.NewMpinEn)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return

	}
	c.JSON(200, gin.H{"response": nil})
}

type discoverAccountReq struct {
	Phone    string `json:"phone" binding:"required"`
	BankCode string `json:"bank_code" binding:"required"`
}

type setMpinReq struct {
	AccountId string `json:"account_id" binding:"required"`
	BankCode  string `json:"bank_code" binding:"required"`
	MpinEn    string `json:"mpin_encrypted" binding:"required"`
}

type changeMpinReq struct {
	AccountId string `json:"account_id" binding:"required"`
	BankCode  string `json:"bank_code" binding:"required"`
	OldMpinEn string `json:"old_mpin_encrypted" binding:"required"`
	NewMpinEn string `json:"new_mpin_encrypted" binding:"required"`
}

type getBalanceReq struct {
	AccountId string `json:"account_id" binding:"required"`
	BankCode  string `json:"bank_code" binding:"required"`
	MpinEn    string `json:"mpin_encrypted" binding:"required"`
}
