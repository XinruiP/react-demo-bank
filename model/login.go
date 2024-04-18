package model

import (
	"demo/database"
	"errors"
	"github.com/golang-jwt/jwt"
	"time"
)

type User struct {
	Id       int    `db:"id"`
	Name     string `db:"name"`
	Password string `db:"password"`
}
type UseRes struct {
	UserId     int      `json:"user_id"`
	Name       string   `json:"name"`
	Token      string   `json:"token"`
	Department []string `json:"department"`
	Role       string   `json:"role"`
}

func Login(username, password string) (UseRes, error) {
	var user User
	var userRes UseRes
	db := database.Mysql
	// 查询用户
	if err := db.Raw("SELECT id, name, password, role_id,user_id FROM login WHERE name = ?", username).Scan(&user).Error; err != nil {
		return UseRes{}, err
	}
	// 验证密码
	if user.Password != password {
		// 密码错误
		return UseRes{}, errors.New("wrong password")
	}

	// 查询用户角色
	var role string
	var department []string
	var name string
	if err := db.Raw("SELECT role.role FROM login LEFT JOIN role ON login.role_id = role.id WHERE login.id = ?", user.Id).Scan(&role).Error; err != nil {
		return UseRes{}, err
	}
	if err := db.Raw(`
		SELECT
			department.name
		FROM
			login
			LEFT JOIN users ON login.user_id = users.id
			LEFT JOIN user_son ON users.id = user_son.user_id
			LEFT JOIN department ON user_son.department_id = department.id 
		WHERE
			login.id = ?
		`, user.Id).Scan(&department).Error; err != nil {
		return UseRes{}, err
	}
	if err := db.Raw("SELECT users.name FROM login LEFT JOIN users ON login.user_id = users.id WHERE login.id = ?", user.Id).Scan(&name).Error; err != nil {
		return UseRes{}, err
	}
	// 生成JWT
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Name
	claims["role"] = role
	claims["password"] = user.Password
	claims["userid"] = user.Id
	claims["department"] = department
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix() // Token过期时间为一月

	tokenString, err := token.SignedString([]byte("akdfjkdasjpkhffdjahios"))
	if err != nil {
		return UseRes{}, err
	}
	userRes.UserId = user.Id
	userRes.Name = name
	userRes.Token = tokenString
	userRes.Role = role
	userRes.Department = department
	return userRes, nil
}
