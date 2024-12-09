package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gochat/internal/schema"
	"gochat/internal/services"
	"net/http"
)

type CreateAccountRequest struct {
	Name string `form:"name" json:"name" binding:"required"`
}

type CreateAccountDomainRequest struct {
	AccountID string `form:"accountId" json:"accountId" binding:"required"`
	Domain    string `form:"domain" json:"domain"`
}

type AccountHandlers struct {
	accountService services.AccountService
}

func NewAccountHandlers(as *services.AccountService) *AccountHandlers {
	return &AccountHandlers{
		accountService: *as,
	}
}

func (h *AccountHandlers) GetAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.Param("id")

		account, err := h.accountService.Get(c, accountID)
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusNotFound)
		}

		c.JSON(http.StatusOK, gin.H{
			"account": account,
		})

	}
}

func (h *AccountHandlers) CreateAccount() gin.HandlerFunc {
	return func(c *gin.Context) {

		var params CreateAccountRequest
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var accountData schema.CreateAccountParams
		accountData.ID = uuid.New().String()
		accountData.Name = params.Name

		newAccount, err := h.accountService.Create(c, accountData)

		if err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusNotFound)
		}

		c.JSON(http.StatusOK, gin.H{
			"account": newAccount,
		})

	}
}

func (h *AccountHandlers) AddDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params CreateAccountDomainRequest
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		newDomain, err := h.accountService.CreateAccountDomain(c, schema.CreateAccountDomainParams{Domain: params.Domain, Account: params.AccountID})
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"domain": newDomain,
		})
	}
}
