package user

import (
	"net/http"

	"github.com/casdoor/casdoor/internal/object"
	"github.com/casdoor/casdoor/internal/store"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	userStore *store.UserStore
}

func New(userStore *store.UserStore) *Handler {
	return &Handler{userStore: userStore}
}

func (h *Handler) GetUsers(g *gin.Context) {
	owner := g.GetString("owner")
	g.JSON(http.StatusOK, h.userStore.GetUsers(owner))
}

func (h *Handler) GetUser(g *gin.Context) {
	id := g.GetString("id")
	u, err := h.userStore.GetUser(id)
	if err != nil {
		_ = g.Error(err)
		return
	}
	g.JSON(http.StatusOK, u)
}

func (h *Handler) UpdateUser(g *gin.Context) {
	id := g.GetString("id")

	var user object.User
	err := g.BindJSON(&user)
	if err != nil {
		panic(err)
	}

	ok, err := h.userStore.UpdateUser(id, &user)
	if err != nil {
		_ = g.Error(err)
		return
	}
	g.JSON(http.StatusOK, ok)
}

func (h *Handler) AddUser(g *gin.Context) {
	var user object.User
	err := g.BindJSON(&user)
	if err != nil {
		panic(err)
	}
	ok, err := h.userStore.AddUser(&user)
	if err != nil {
		_ = g.Error(err)
		return
	}
	g.JSON(http.StatusOK, ok)
}

func (h *Handler) DeleteUser(g *gin.Context) {
	var user object.User
	err := g.BindJSON(&user)
	if err != nil {
		_ = g.Error(err)
		return
	}

	ok, err := h.userStore.DeleteUser(&user)
	if err != nil {
		_ = g.Error(err)
		return
	}
	g.JSON(http.StatusOK, ok)
}
