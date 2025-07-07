package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

type Config struct {
	SSHHost string `yaml:"ssh_host"`
	SSHPort int    `yaml:"ssh_port"`
	SSHUser string `yaml:"ssh_user"`
	SSHKey  string `yaml:"ssh_key"`
	WebPort int    `yaml:"web_port"`
}

var config Config

func loadConfig() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("配置文件格式错误: %v", err)
	}
}

func sshAsk(question string) (string, error) {
	key, err := ioutil.ReadFile(config.SSHKey)
	if err != nil {
		return "", fmt.Errorf("无法读取私钥: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("无法解析私钥: %w", err)
	}
	sshConfig := &ssh.ClientConfig{
		User:            config.SSHUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr := fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return "", fmt.Errorf("SSH连接失败: %w", err)
	}
	defer client.Close()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY 环境变量未设置")
	}
	envPath := "/home/ubuntu/.nvm/versions/node/v22.17.0/bin:/home/ubuntu/.npm-global/bin:/home/ubuntu/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	geminiPath := "/home/ubuntu/.npm-global/bin/gemini"

	// 创建执行会话
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("无法创建SSH会话: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	session.Setenv("GEMINI_API_KEY", apiKey)
	session.Setenv("PATH", envPath)
	session.Setenv("NVM_DIR", "/home/ubuntu/.nvm")

	cmd := fmt.Sprintf(`GEMINI_API_KEY="%s" %s -p "%s"`, apiKey, geminiPath, question)
	log.Printf("[INFO] 执行命令: %s", cmd)
	if err := session.Run(cmd); err != nil {
		output := stdout.String()
		errorOutput := stderr.String()
		log.Printf("[ERROR] 命令执行失败: %v", err)
		log.Printf("[ERROR] stdout: %s", output)
		log.Printf("[ERROR] stderr: %s", errorOutput)
		return output, fmt.Errorf("命令执行失败: %w, stderr: %s", err, errorOutput)
	}
	output := stdout.String()
	log.Printf("[DEBUG] 命令输出: %s", output)
	return output, nil
}

func main() {
	loadConfig()
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	// 静态页面
	r.StaticFile("/", "./index.html")
	r.POST("/ask", func(c *gin.Context) {
		var req struct {
			Question string `json:"question"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] 参数错误: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		log.Printf("[INFO] 收到提问: %s", req.Question)
		answer, err := sshAsk(req.Question)

		// ==== 新增：记录问答到日志 ====
		go func(q, a string) {
			f, ferr := os.OpenFile("qa.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if ferr == nil {
				logLine := fmt.Sprintf("%s\nQ: %s\nA: %s\n\n", time.Now().Format("2006-01-02 15:04:05"), q, a)
				f.WriteString(logLine)
				f.Close()
			}
		}(req.Question, answer)
		// ==== 新增结束 ====

		if err != nil {
			log.Printf("[ERROR] 执行命令出错: %v, 输出: %s", err, answer)
			c.JSON(http.StatusOK, gin.H{"error": err.Error(), "output": answer})
			return
		}
		log.Printf("[INFO] 返回答案: %s", answer)
		c.JSON(http.StatusOK, gin.H{"answer": answer})
	})
	addr := fmt.Sprintf("0.0.0.0:%d", config.WebPort)
	fmt.Printf("服务已启动: http://%s\n", addr)
	r.Run(addr)
}
