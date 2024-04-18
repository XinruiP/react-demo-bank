package api

import (
	"demo/model"
	"fmt"
	camunda_client_go "github.com/citilinkru/camunda-client-go/v3"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ChooseUser struct {
	CompanyId int `json:"company_id"`
	UserId    int `json:"user_id"`
}

type Response struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data"`
}
type ApplyInstanceId struct {
	ProcessInstanceId    string `json:"process_instance_id"`
	IdOfFinishInstanceId string `json:"id_of_finish_instance_id"`
}
type Upload struct {
	IdOfTask string `json:"id_of_task"`
	UserId   int    `json:"user_id"`
}
type FinishTask struct {
	ProcessInstanceId string `json:"process_instance_id"`
	IsAgree           string `json:"is_agree"`
	Result            string `json:"result"`
}

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type StartProcessRequest struct {
	Variables map[string]camunda_client_go.Variable `json:"variables"`
}
type FinishProcessTask struct {
	Id                    string `json:"id"`
	ProcessDefinitionKey  string `json:"processDefinitionKey"`
	ProcessDefinitionId   string `json:"processDefinitionId"`
	ProcessInstanceId     string `json:"processInstanceId"`
	ExecutionId           string `json:"executionId"`
	CaseDefinitionKey     string `json:"caseDefinitionKey"`
	CaseDefinitionId      string `json:"caseDefinitionId"`
	CaseInstanceId        string `json:"caseInstanceId"`
	CaseExecutionId       string `json:"caseExecutionId"`
	ActivityInstanceId    string `json:"activityInstanceId"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	DeleteReason          string `json:"deleteReason"`
	Owner                 string `json:"owner"`
	Assignee              string `json:"assignee"`
	StartTime             string `json:"startTime"`
	EndTime               string `json:"endTime"`
	Duration              int64  `json:"duration"`
	TaskDefinitionKey     string `json:"taskDefinitionKey"`
	Priority              int64  `json:"priority"`
	Created               string `json:"created"`
	Due                   string `json:"due"`
	ParentTaskId          string `json:"parentTaskId"`
	FollowUp              string `json:"followUp"`
	TenantId              string `json:"tenantId"`
	RemovalTime           string `json:"removalTime"`
	RootProcessInstanceId string `json:"rootProcessInstanceId"`
}

func APIResource(status int, objects interface{}, msg string) (resp *Response) {
	resp = &Response{Code: status, Data: objects, Msg: msg}
	return
}

// SendApproval 发起申请
func SendApproval(ctx echo.Context) error {
	//var choose ChooseUser
	var req StartProcessRequest
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "解析body错误"))
	}
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := claims["userid"].(float64)
	userIdStr := strconv.FormatFloat(userId, 'f', -1, 64)
	department := claims["department"].([]interface{})
	var processType = "demo1"
	//variables := make(map[string]camunda_client_go.Variable)
	Sub := make(map[string]camunda_client_go.Variable)
	for key, variable := range req.Variables {
		if key == "two_grade" {
			Sub["two_grade"] = variable
		}
	}
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	for _, s := range department {
		if s == "市场综合室+技术商务室" {
			result, err := client.ProcessDefinition.StartInstance(
				camunda_client_go.QueryProcessDefinitionBy{Key: &processType},
				camunda_client_go.ReqStartInstance{
					Variables: &req.Variables,
				},
			)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "发起流程失败"))
			}
			list, err := client.UserTask.GetList(&camunda_client_go.UserTaskGetListQuery{
				Assignee:          userIdStr,
				ProcessInstanceId: result.Id,
			})
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "获取实例id失败"))
			}
			var applyInstance string
			for _, task := range list {
				applyInstance = task.Id
			}
			err = client.UserTask.Complete(applyInstance, camunda_client_go.QueryUserTaskComplete{
				Variables: Sub,
			})
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "发起申请失败"))
			}
			return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, nil, "发起申请成功"))
		}
	}
	return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, nil, "不属于一级单位"))
}

// QueryWaitToDo 查询待办
func QueryWaitToDo(ctx echo.Context) error {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := claims["userid"].(float64)
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	list, err := client.UserTask.GetList(&camunda_client_go.UserTaskGetListQuery{
		Assignee: strconv.FormatFloat(userId, 'f', -1, 64),
	})
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "查询个人待完成任务失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, list, "查询个人待完成任务成功"))
}

// 上传文件
func UploadFile(ctx echo.Context) error {
	// 从请求中获取文件
	file, err := ctx.FormFile("file")
	processId := ctx.FormValue("processInstanceId")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "获取文件失败"))
	}
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "打开文件失败"))
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create("/Users/ruipan/file/" + file.Filename)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "创建文件失败"))
	}
	defer dst.Close()

	// 将源文件内容拷贝到目标文件
	if _, err = io.Copy(dst, src); err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "保存文件失败"))
	}
	// 构建文件在服务器上的 URL
	// 获取服务器的IP地址
	host, _, err := net.SplitHostPort(ctx.Request().Host)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "处理ip地址错误"))
	}
	serverURL := fmt.Sprintf("http://%s:9999/%s", host, file.Filename)
	err = model.InsertFile(serverURL, processId)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "文件URL存入数据库失败"))
	}
	// 解析body
	// 解析表单字段
	idOfTask := ctx.FormValue("id_of_task")
	userId := ctx.FormValue("user_id")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "解析body失败"))
	}
	// 完成待办
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	variables := make(map[string]camunda_client_go.Variable)
	variables["two_grade_leader"] = camunda_client_go.Variable{
		Type:  "String",
		Value: userId,
	}
	err = client.UserTask.Complete(idOfTask, camunda_client_go.QueryUserTaskComplete{
		Variables: variables,
	})
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "完成待办任务失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, nil, "查询个人待完成任务成功"))
}

// 查询二级单位
func QueryTwoGrades(ctx echo.Context) error {
	company, err := model.QueryTwoGradeCompany()
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "查询失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, company, "查询成功"))
}

// 查询二级单位负责人
func QueryTwoGradeUser(ctx echo.Context) error {
	param := ctx.QueryParam("companyId")
	atoi, err := strconv.Atoi(param)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "转换int失败"))
	}
	users, err := model.QueryTwoGradeUserId(atoi)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "查询失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, users, "查询成功"))
}

// 查询二级单位领导人
func QueryTwoGradeLeader(ctx echo.Context) error {
	param := ctx.QueryParam("processInstanceId")
	var departmentId string
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	query := map[string]string{
		"processInstanceId": param,
	}
	instances, err := client.History.GetVariableInstanceList(query)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "查询失败"))
	}
	for _, instance := range instances {
		if instance.Name == "department_id" {
			if i, ok := instance.Value.(string); ok {
				departmentId = i
			} else {
				// 如果类型断言失败，可以处理错误或者采取其他适当的措施
				return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, nil, "部门ID类型断言失败"))
			}
		}
	}
	atoi, err := strconv.Atoi(departmentId)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "转换int错误"))
	}
	leader, err := model.QueryTwoGradeLeader(atoi)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "查询失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, leader, "查询成功"))
}

// 领导审批
func LeaderExamine(ctx echo.Context) error {
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	var processInstance FinishTask
	variables := make(map[string]camunda_client_go.Variable)
	err := ctx.Bind(&processInstance)
	variables["isagree"] = camunda_client_go.Variable{
		Type:  "String",
		Value: processInstance.IsAgree,
	}
	variables["result"] = camunda_client_go.Variable{
		Type:  "String",
		Value: processInstance.Result,
	}
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "解析body失败"))
	}
	err = client.UserTask.Complete(processInstance.ProcessInstanceId, camunda_client_go.QueryUserTaskComplete{
		Variables: variables,
	})
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "审核失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, nil, "审核成功"))
}

// 查询已经完成的任务
func QueryFinishTask(ctx echo.Context) error {
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := claims["userid"].(float64)
	query := map[string]string{
		"taskAssignee": strconv.FormatFloat(userId, 'f', -1, 64),
		"finished":     "true",
	}
	instances, err := client.History.GetTaskList(query)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "查询失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, instances, "查询成功"))
}

// 查询流程变量
func QueryVariables(ctx echo.Context) error {
	param := ctx.QueryParam("processInstanceId")
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	query := map[string]string{
		"processInstanceId": param,
	}
	instances, err := client.History.GetVariableInstanceList(query)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, nil, "查询流程变量失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, instances, "查询流程变量成功"))
}

// 查看文件
func QueryFile(ctx echo.Context) error {
	param := ctx.QueryParam("processInstanceId")
	file, err := model.QueryFile(param)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "查询文件路径失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, file, "查询文件路径成功"))
}

// 获取活动实例
func QueryActivity(ctx echo.Context) error {
	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "admin",
		ApiPassword: "123456",
		Timeout:     time.Second * 10,
	})
	param := ctx.QueryParam("processInstanceId")
	instance, err := client.ProcessInstance.GetActivityInstance(param)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, APIResource(http.StatusBadRequest, err, "查询活动失败"))
	}
	return ctx.JSON(http.StatusOK, APIResource(http.StatusOK, instance, "查询活动成功"))
}
