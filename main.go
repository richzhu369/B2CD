package main

import (
	"fmt"
)

const downloadURL = "https://jenkins-buildpackage.s3.ap-east-1.amazonaws.com/"

//const downloadURL = "http://127.0.0.1:8080/"

func main() {
	// 获取执行参数
	packageName, appName, envName, md5Value, action, servers, checkMD5 := getArgs()

	fmt.Println(envName)
	switch *action {
	case "deploy":
		Deploy(packageName, appName, md5Value, servers, checkMD5)
	case "rollback":
		Rollback()
	case "restart":
		Restart(*appName, *servers)
	default:
		fmt.Println("选项错误，目前只支持deploy、rollback、restart")
	}

}
