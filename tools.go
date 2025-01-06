package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
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

// executeSSHCommand ssh到服务器执行命令，30秒超时
func executeSSHCommand(server, command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh", "-p", "10086", "root@"+server, command)
	log.Println("Executing command: ", cmd.String())
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("command timed out")
	}
	if err != nil {
		log.Printf("Error executing command: %v", err)
		return fmt.Errorf("failed to execute command: %s, output: %s, error: %w", command, output, err)
	}
	log.Println("Command executed successfully")
	log.Println("Command output: ", string(output))
	return nil
}

// 部署应用
func deployApp(srcPath, destPath, serverAddress, sshUser, sshKey string) error {
	// 实现应用部署逻辑
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

func getArgs() (packageName, appName, envName, md5Value, action, servers *string, checkMD5 *bool) {
	// Define command line flags
	packageName = flag.String("packageName", "", "包名")
	appName = flag.String("appName", "", "Name of the application")
	envName = flag.String("envName", "", "Environment name")
	md5Value = flag.String("md5Value", "", "Expected MD5 value of the application package")
	checkMD5 = flag.Bool("checkMD5", false, "Whether to check the MD5 value of the application package")
	action = flag.String("action", "", "Action to perform")
	servers = flag.String("servers", "", "Comma-separated list of servers to deploy the application to") //todo 处理多个地址，转成列表方便部署的时候循环执行远程命令

	flag.Parse()

	// 检查参数是否为空
	if *packageName == "" || *appName == "" || *envName == "" {
		fmt.Println("Missing required flags")
		flag.Usage()
		os.Exit(1)
	}

	return packageName, appName, envName, md5Value, action, checkMD5
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
