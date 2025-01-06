package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Deploy 部署应用
func Deploy(packageName, appName, md5Value, serverAddrs *string, checkMD5 *bool) {
	// 创建工作目录
	dir, err := createWorkingDir("app")
	if err != nil {
		log.Fatal("Error creating working directory:", err)
		return
	}
	// 切换当前目录为解压目录
	err = os.Chdir(dir)

	// 下载应用包
	err = downloadFile(downloadURL, *packageName)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}

	// MD5校验
	if *checkMD5 {
		if !verifyMD5("./", *md5Value) {
			log.Fatal("传入的MD5和计算压缩包的MD5不一致")
			return
		}
	}

	// 解压应用包
	err = extractTarGz(*packageName, "./")
	if err != nil {
		log.Fatal("Error 解压文件错误:", err)
		return
	}

	// 获取解压后的文件路径,路径运行时传入的包名去掉.tar.gz后缀
	workingPath := filepath.Join("./", (*packageName)[:len(*packageName)-7], "/")

	// 计算应用MD5并与.md5sum文件中的值进行比较
	appFilePath := filepath.Join(workingPath, *appName)
	md5sumFilePath := filepath.Join(workingPath, *appName+".md5sum")
	if !verifyAppMD5(appFilePath, md5sumFilePath) {
		log.Fatal("应用MD5和.md5sum文件中的值不一致")
		return
	}
	log.Println("应用MD5校验通过,与.md5sum文件中的值一致")

	// 部署应用
	deployPath := fmt.Sprintf("/data/app/%s/release", *appName)
	err = deployApp(workingPath, deployPath, *serverAddrs)
	if err != nil {
		fmt.Println("Error deploying application:", err)
		return
	}

	//// 更新软连接
	//currentLink := fmt.Sprintf("/data/app/%s/current", appName)
	//err = updateSymlink(deployPath, currentLink, serverAddress, sshUser, sshKey)
	//if err != nil {
	//	fmt.Println("Error updating symlink:", err)
	//	return
	//}

	// 重启应用
}

// Restart 重启应用
func Restart() {

}

// Rollback 回滚应用
func Rollback() {

}
