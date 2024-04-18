package model

import "demo/database"

type Company struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type Users struct {
	Id   int    `json:"id" gorm:"id"`
	Name string `json:"name" gorm:"name"`
}

// QueryTwoGradeCompany 查询二级单位
func QueryTwoGradeCompany() ([]Company, error) {
	db := database.Mysql
	var company []Company
	sql := "select id,name from department where parent_id = 1"
	if err := db.Raw(sql).Scan(&company).Error; err != nil {
		return nil, err
	}
	return company, nil
}

// QueryTwoGradeUserId 查询二级单位负责人
func QueryTwoGradeUserId(companyId int) ([]Users, error) {
	db := database.Mysql
	var users []Users
	sql := `
		SELECT
			users.id,
			users.name
		FROM
			department
			LEFT JOIN user_son ON department.id = user_son.department_id
			LEFT JOIN users ON user_son.user_id = users.id
			LEFT JOIN role ON users.role_id = role.id 
		WHERE
			department.id = ?
			AND role.role = '负责人'
		`
	if err := db.Raw(sql, companyId).Scan(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// InsertFile 插入文件
func InsertFile(url string, processId string) error {
	db := database.Mysql
	sql := "INSERT INTO `upload` (file,process_id) VALUES (?,?)"
	if err := db.Exec(sql, url, processId).Error; err != nil {
		return err
	}
	return nil
}

// 查询二级单位领导
func QueryTwoGradeLeader(departmentId int) ([]Users, error) {
	db := database.Mysql
	var users []Users
	sql := `
				SELECT
					users.id,
					users.name 
				FROM
					department
					LEFT JOIN user_son ON department.id = user_son.department_id
					LEFT JOIN users ON user_son.user_id = users.id
					LEFT JOIN role ON users.role_id = role.id 
				WHERE
					department.id = ?
				AND role.role = '责任领导'
			`
	if err := db.Raw(sql, departmentId).Scan(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// 查看文件
func QueryFile(processId string) (string, error) {
	db := database.Mysql
	var file string
	sql := "select file from upload where process_id = ?"
	if err := db.Raw(sql, processId).Scan(&file).Error; err != nil {
		return "", err
	}
	return file, nil
}
