package main

import (
	demo "demo/api"
	"demo/database"
	"fmt"
	camunda_client_go "github.com/citilinkru/camunda-client-go/v3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"os"
	"time"
)

func main() {
	// 初始化数据库连
	database.InitMysql()
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	// 部署流程
	file, err := os.Open("/Users/ruipan/Desktop/demo/bpmn/demo.bpmn")
	if err != nil {
		fmt.Printf("Error read file: %s\n", err)
		return
	}
	_, err = client.Deployment.Create(camunda_client_go.ReqDeploymentCreate{
		DeploymentName: "Process_14xg0vo",
		Resources: map[string]interface{}{
			"demo.bpmn": file,
		},
	})
	if err != nil {
		panic(err)
	}
	e := echo.New()
	// CORS跨域
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		//http://localhost:8080（前端web项目的IP及端口号） 浏览器访问存在跨区问题，需要对此开放
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAcceptEncoding, echo.HeaderAuthorization},
	}))
	api := e.Group("/api/v1")
	// 登录接口不经过中间件
	e.POST("/login", demo.Login)
	// 中间件：用于处理跨域请求等
	api.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte("akdfjkdasjpkhffdjahios"),
		TokenLookup: "header:" + echo.HeaderAuthorization,
		ContextKey:  "user",
	}))
	api.POST("/sendApproval", demo.SendApproval)
	api.GET("/queryWaitToDo", demo.QueryWaitToDo)
	api.POST("/upload", demo.UploadFile)
	api.GET("/queryTwoGradeCompany", demo.QueryTwoGrades)
	api.GET("/queryTwoGradeUser", demo.QueryTwoGradeUser)
	api.GET("/queryTwoGradeLeader", demo.QueryTwoGradeLeader)
	api.POST("/examine", demo.LeaderExamine)
	api.GET("/queryFinished", demo.QueryFinishTask)
	api.GET("/queryVariables", demo.QueryVariables)
	api.GET("/queryFileUrl", demo.QueryFile)
	api.GET("/queryActivity", demo.QueryActivity)
	e.Logger.Fatal(e.Start(":1323"))
}
