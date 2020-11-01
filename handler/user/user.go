package user

import (
	"net/http"

	"github.com/casdoor/casdoor/object"
	"github.com/gin-gonic/gin"
)

func GetUsers(g *gin.Context) {
	owner := g.GetString("owner")
	g.JSON(http.StatusOK, object.GetUsers(owner))
}

func GetUser(g *gin.Context) {
	id := g.GetString("id")
	g.JSON(http.StatusOK, object.GetUser(id))
}

func UpdateUser(g *gin.Context) {
	id := g.GetString("id")

	var user object.User
	err := g.BindJSON(&user)
	if err != nil {
		panic(err)
	}

	g.JSON(http.StatusOK, object.UpdateUser(id, &user))
}

func AddUser(g *gin.Context) {
	var user object.User
	err := g.BindJSON(&user)
	if err != nil {
		panic(err)
	}

	g.JSON(http.StatusOK, object.AddUser(&user))
}

func DeleteUser(g *gin.Context) {
	var user object.User
	err := g.BindJSON(&user)
	if err != nil {
		panic(err)
	}

	g.JSON(http.StatusOK, object.DeleteUser(&user))
}
