package api

import (
	"demo/model"
	"github.com/labstack/echo/v4"
	"net/http"
)

// User 模拟用户数据
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c echo.Context) error {
	var user User
	err2 := c.Bind(&user)
	if user.Username == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, nil, "用户名或者密码为空"))
	}
	if err2 != nil {
		return c.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err2, "解析body错误"))
	}

	login, err2 := model.Login(user.Username, user.Password)
	if err2 != nil {
		return c.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err2, "登陆失败"))
	}
	return c.JSON(http.StatusOK, APIResource(http.StatusOK, login, "登陆成功"))
}
