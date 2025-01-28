package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ssh到服务器执行命令
func executeSSHCommand(server, command string) error {
	cmd := exec.Command("ssh", "-p", "55888", "b2om@"+server, command)
	log.Println("执行命令: ", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error 执行命令: %v", err)
		return fmt.Errorf("执行命令失败: %s, output: %s, error: %w", command, output, err)
	}
	log.Println("Command executed successfully")
	log.Println("Command output: ", string(output))
	return nil
}

// 拷贝文件到服务器
func copyFileToServer(server, src, dest string) error {
	dir, err2 := os.Getwd()
	if err2 != nil {
		log.Fatal("获取当前路径失败：", err2)
	}
	log.Println("当前路径：", dir)

	command := fmt.Sprintf("scp -r -P 55888 %s/* b2om@%s:%s", src, server, dest)
	cmd := exec.Command("sh", "-c", command)

	log.Println("执行命令: ", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error 执行命令: %v", err)
		return fmt.Errorf("执行命令失败: %s, output: %s, error: %w", cmd.String(), output, err)
	}
	log.Println("执行命令成功")
	log.Println("命令: ", string(output))
	return nil
}

// 处理并获得应用文件夹的名字
func getAppDirName(packageName string) string {
	reduceTail := strings.TrimSuffix(packageName, ".tar.gz")
	parts := strings.Split(reduceTail, "_")
	log.Println("包名：", parts)
	log.Println(len(parts))
	if len(parts) > 2 {
		extractedString := strings.Join(parts[2:], "_")
		log.Println("应用文件夹名字：", extractedString)
		return extractedString
	} else {
		log.Fatal("包名格式错误")
		return ""
	}
}

// 检查systemd是否存在
func checkSystemd(appName, workingPath, serverAddress string) error {
	log.Println("检查systemd是否存在")
	// 检查systemd是否存在
	err := executeSSHCommand(serverAddress, "systemctl status "+appName)
	if err != nil {
		log.Println("systemd 不存在，创建")
		// 创建systemd文件
		systemdFile := fmt.Sprintf(`[Unit]`+"\n"+`Description=%s`+"\n"+`After=network-online.target remote-fs.target nss-lookup.target`+"\n"+"Wants=network-online.target"+"\n"+"\n"+`[Service]`+"\n"+`Type=simple`+"\n"+"WorkingDirectory=/data/app/%s/current/"+"\n"+`ExecStart=/data/app/%s/current/%s`+"\n"+"KillSignal=SIGTERM"+"\n"+"SendSIGKILL=no"+"\n"+"SuccessExitStatus=0"+"\n"+"LimitNOFILE=200000"+"\n"+`Restart=always`+"\n"+"\n"+`[Install]`+"\n"+`WantedBy=multi-user.target`, appName, appName, appName, appName)
		systemdFilePath := filepath.Join(workingPath, appName+".service")
		err = os.WriteFile(systemdFilePath, []byte(systemdFile), 0644)
		if err != nil {
			log.Fatal("创建systemd文件失败：", err)
			return err
		}

		// 上传systemd文件
		err = copySystemdFileToServer(serverAddress, systemdFilePath, "/data/app/"+appName+"/release/")
		if err != nil {
			log.Fatal("上传systemd文件失败：", err)
			return err
		}
	}
	return nil
}

// 部署应用
func deployToServer(appName, srcPath, destPath, serverAddress string) error {
	log.Println("开始部署")

	// 0. 拆分服务器地址，如果是多个以逗号分隔
	ips := strings.Split(serverAddress, ",")
	for _, ip := range ips {
		log.Println("部署到服务器：", ip)
		// 1. 创建远程目录
		err := executeSSHCommand(ip, fmt.Sprintf("mkdir -pv %s", destPath))
		if err != nil {
			log.Fatal("创建远程目录失败：", err)
		}
		// 清空一次目录，防止相同包重复部署时遇到错误
		err = executeSSHCommand(ip, fmt.Sprintf("rm -r -f %s/*", destPath))
		if err != nil {
			log.Fatal("清空远程目录失败：", err)
		}

		// 2. 上传文件
		err = copyFileToServer(ip, srcPath, destPath)
		if err != nil {
			log.Println("上传文件失败：", err)
			return fmt.Errorf("上传文件失败：%w", err)
		}

		// 检测systemd 是否存在，如果不存在就创建并执行 systemctl daemon-reload
		err = checkSystemd(appName, srcPath, ip)
		if err != nil {
			fmt.Println("Error checking systemd:", err)
			return nil
		} else {
			// 拷贝systemd文件到/usr/lib/systemd/system/
			err := executeSSHCommand(ip, fmt.Sprintf("sudo cp -f /data/app/%s/release/%s.service /usr/lib/systemd/system/", appName, appName))
			if err != nil {
				log.Fatal("在服务器中拷贝systemd文件失败：", err)
				return err
			}
			// 重载systemd
			err = executeSSHCommand(ip, "sudo systemctl daemon-reload")
			if err != nil {
				log.Fatal("重载systemd失败：", err)
				return err
			}
		}

		// 3. 更改软连接，指向新版本
		err = executeSSHCommand(ip, fmt.Sprintf("ln -sfn %s %s", destPath, "/data/app/"+appName+"/current"))
		if err != nil {
			log.Fatal("更改软连接失败：", err)
		}

		// 4. 检测当前有几个历史版本，如果超过5个，删除最旧的版本
		err = executeSSHCommand(ip, fmt.Sprintf("cd /data/app/%s/release && ls -t | tail -n +6 | xargs rm -rf", appName))
		// 5. 重启应用
		err = executeSSHCommand(ip, fmt.Sprintf("sudo systemctl restart %s", appName))
	}

	return nil
}

// 拷贝systemd文件到服务器
func copySystemdFileToServer(server, src, dest string) error {
	command := fmt.Sprintf("scp -r -P 55888 %s b2om@%s:%s", src, server, dest)
	cmd := exec.Command("sh", "-c", command)

	log.Println("执行命令: ", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error 执行命令: %v", err)
		return fmt.Errorf("执行命令失败: %s, output: %s, error: %w", cmd.String(), output, err)
	}
	log.Println("执行命令成功")
	log.Println("命令: ", string(output))
	return nil
}

// 创建打包目录
func createWorkingDir(appName string) (string, error) {
	// 在当前目录下创建一个临时目录
	dir, err := os.MkdirTemp(".", appName)
	if err != nil {
		log.Fatal("Error creating temporary directory:", err)
		return "", err
	}
	//defer os.RemoveAll(dir)
	return dir, nil
}

func getArgs() (packageName, appName, envName, md5Value, action, servers, checkMD5 *string) {
	// Define command line flags
	packageName = flag.String("packageName", "", "包名")
	appName = flag.String("appName", "", "Name of the application")
	envName = flag.String("envName", "", "Environment name")
	md5Value = flag.String("md5Value", "", "Expected MD5 value of the application package")
	checkMD5 = flag.String("checkMD5", "", "Whether to check the MD5 value of the application package")
	action = flag.String("action", "", "Action to perform")
	servers = flag.String("servers", "", "Comma-separated list of servers to deploy the application to") //todo 处理多个地址，转成列表方便部署的时候循环执行远程命令

	flag.Parse()

	// 检查参数是否为空
	if *packageName == "" || *appName == "" || *envName == "" {
		fmt.Println("Missing required flags")
		flag.Usage()
		os.Exit(1)
	}

	return packageName, appName, envName, md5Value, action, servers, checkMD5
}

func getFilenameFromURL(url string) string {
	//尝试从Content-Disposition 头部获取文件名
	if parts := strings.Split(url, "/"); len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if lastPart != "" {
			return lastPart
		}
	}
	return ""
}

// 下载文件
func downloadFile(downloadURL, packageName string) error {
	timeout := time.Duration(5 * time.Minute)
	client := http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			fmt.Println("重定向到：", req.URL)
			return nil // 允许重定向
		},
	}

	resp, err := client.Get(downloadURL + packageName)
	if err != nil {
		return fmt.Errorf("获取 URL 失败：%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 错误：%d %s", resp.StatusCode, resp.Status)
	}

	// 提取文件名
	filename := getFilenameFromURL(downloadURL + packageName)
	if filename == "" {
		filename = "downloaded_file"
	}

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败：%w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}

	fmt.Println("文件下载成功，保存为：", filename)
	return nil
}

// 校验MD5
func verifyMD5(filepath, expectedMD5 string) bool {
	fmt.Println("校验文件：", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		return false
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false
	}

	calculatedMD5 := hex.EncodeToString(hash.Sum(nil))
	return calculatedMD5 == expectedMD5
}

// 解压tar.gz文件
func extractTarGz(gzipPath, destPath string) error {
	file, err := os.Open(gzipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destPath, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		default:
			log.Printf("Unable to untar type: %c in file %s", header.Typeflag, header.Name)
		}
	}
	return nil
}

// 校验应用MD5
func verifyAppMD5(appFilePath, md5sumFilePath string) bool {
	fmt.Println("校验应用MD5", appFilePath, md5sumFilePath)
	appMD5, err := calculateMD5(appFilePath)
	if err != nil {
		return false
	}

	md5sumContent, err := os.ReadFile(md5sumFilePath)
	if err != nil {
		return false
	}
	fmt.Println("MD5SUM文件内容：", string(md5sumContent))

	expectedMD5 := strings.TrimSpace(string(md5sumContent))
	fmt.Println("应用MD5：", appMD5)
	return appMD5 == expectedMD5
}

// 计算文件MD5
func calculateMD5(filepath string) (string, error) {
	log.Println("计算文件MD5：", filepath)

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal("读取应用文件错误:", err)
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal("计算 MD5 错误:", err)
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
