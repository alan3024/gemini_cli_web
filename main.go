package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/ssh"
)

//go:embed font/chinese.ttf
var embeddedFont []byte

// --- 自定义主题，用于设置全局字体和字号 ---
type myTheme struct {
	fyne.Theme
}

var (
	chineseFontResource fyne.Resource
	defaultFontSize     float32 = 16 // 调整为适中的字号
)

// newMyTheme 初始化自定义主题，并从嵌入的资源中创建字体
func newMyTheme() fyne.Theme {
	log.Println("开始初始化自定义主题...")
	log.Printf("嵌入字体数据大小: %d 字节\n", len(embeddedFont))

	if len(embeddedFont) > 0 {
		chineseFontResource = &fyne.StaticResource{
			StaticName:    "chinese.ttf",
			StaticContent: embeddedFont,
		}
		log.Printf("成功创建字体资源: %s\n", chineseFontResource.Name())
	} else {
		log.Println("警告: 嵌入的字体数据为空，将使用系统默认字体")
		return theme.DefaultTheme()
	}

	customTheme := &myTheme{Theme: theme.DefaultTheme()}
	log.Printf("自定义主题初始化完成，默认字号设置为: %.1f\n", defaultFontSize)
	return customTheme
}

func (t *myTheme) Font(style fyne.TextStyle) fyne.Resource {
	if chineseFontResource != nil {
		return chineseFontResource
	}
	log.Println("警告: 使用默认字体资源")
	return t.Theme.Font(style)
}

// Size 设置界面元素大小
func (t *myTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return defaultFontSize
	case theme.SizeNameHeadingText:
		return defaultFontSize + 4
	case theme.SizeNameSubHeadingText:
		return defaultFontSize + 2
	case theme.SizeNameCaptionText:
		return defaultFontSize - 2
	default:
		return t.Theme.Size(name)
	}
}

// --- SSH配置 ---
const (
	host    = "3.27.226.56"
	port    = "22"
	user    = "ubuntu"
	keyPath = "C:/Users/Administrator/.ssh/yamaxun.pem"
)

// testConnection 测试SSH连接
func testConnection() (string, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("无法读取私钥文件 '%s': %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("无法解析私钥: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	}

	serverAddr := host + ":" + port
	client, err := ssh.Dial("tcp", serverAddr, config)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("无法创建会话: %w", err)
	}
	defer session.Close()

	// 执行简单的echo命令测试
	output, err := session.CombinedOutput("echo '连接测试成功'")
	if err != nil {
		return "", fmt.Errorf("命令执行失败: %w", err)
	}

	return string(output), nil
}

// runDiagnostic 运行详细的环境诊断
func runDiagnostic() (string, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("无法读取私钥文件 '%s': %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("无法解析私钥: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 30,
	}

	serverAddr := host + ":" + port
	client, err := ssh.Dial("tcp", serverAddr, config)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("无法创建会话: %w", err)
	}
	defer session.Close()

	// 详细的诊断脚本
	diagnosticScript := `
echo "=== 环境诊断开始 ===";
echo "当前用户: $(whoami)";
echo "当前目录: $(pwd)";
echo "HOME目录: $HOME";
echo "";

echo "=== Node.js 环境检查 ===";
echo "node版本:";
node -v || echo "node未安装";
echo "";
echo "npm版本:";
npm -v || echo "npm未安装";
echo "";

echo "=== NVM 环境检查 ===";
echo "NVM_DIR=$NVM_DIR";
if [ -d "$HOME/.nvm" ]; then
    echo ".nvm目录存在";
    ls -la $HOME/.nvm;
else
    echo ".nvm目录不存在";
fi
echo "";

echo "=== 尝试加载NVM ===";
export NVM_DIR="$HOME/.nvm";
if [ -s "$NVM_DIR/nvm.sh" ]; then
    . "$NVM_DIR/nvm.sh";
    echo "nvm.sh已加载";
    echo "nvm版本:";
    nvm --version || echo "nvm命令不可用";
else
    echo "nvm.sh不存在或为空";
fi
echo "";

echo "=== PATH环境变量 ===";
echo "$PATH";
echo "";

echo "=== Gemini命令检查 ===";
echo "查找gemini命令:";
which gemini || echo "gemini命令未找到";
echo "";
echo "查找全局npm包:";
npm list -g --depth=0 || echo "无法列出全局包";
echo "";

echo "=== 尝试安装Gemini ===";
echo "正在全局安装@google/generative-ai-cli...";
npm install -g @google/generative-ai-cli || echo "安装失败";
echo "";

echo "=== 重新检查Gemini ===";
echo "重新查找gemini命令:";
which gemini || echo "gemini命令仍未找到";
echo "";

echo "=== 诊断结束 ===";
`

	// 优先用 /bin/bash，如果失败再用 sh
	cmd := fmt.Sprintf("/bin/bash -c %s", strconv.Quote(diagnosticScript))
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		// fallback to sh
		cmd2 := fmt.Sprintf("sh -c %s", strconv.Quote(diagnosticScript))
		output2, err2 := session.CombinedOutput(cmd2)
		if err2 != nil {
			return string(output) + "\n\n" + string(output2), fmt.Errorf("诊断脚本执行失败: %w (sh也失败: %w)", err, err2)
		}
		return string(output2), nil
	}
	return string(output), nil
}

func main() {
	// 设置适中的DPI缩放
	if runtime.GOOS == "windows" {
		os.Setenv("FYNE_SCALE", "1.2")
		os.Setenv("FYNE_FONT_ANTIALIAS", "true")
		log.Println("已设置 FYNE_SCALE=1.2 和字体抗锯齿")
	}

	// 创建应用
	myApp := app.New()
	customTheme := newMyTheme()
	myApp.Settings().SetTheme(customTheme)

	// 创建主窗口
	myWindow := myApp.NewWindow("Gemini 提问客户端")
	myWindow.Resize(fyne.NewSize(650, 500))
	myWindow.SetFixedSize(true)

	// 创建状态显示标签
	statusLabel := canvas.NewText("未测试连接", theme.ErrorColor())
	statusLabel.TextSize = 14
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 创建测试按钮
	testBtn := widget.NewButton("测试连接", nil)
	testBtn.OnTapped = func() {
		testBtn.Disable()
		statusLabel.Color = theme.WarningColor()
		statusLabel.Text = "正在测试连接..."
		statusLabel.Refresh()

		go func() {
			output, err := testConnection()
			if err != nil {
				statusLabel.Color = theme.ErrorColor()
				statusLabel.Text = "连接失败: " + err.Error()
			} else {
				statusLabel.Color = theme.SuccessColor()
				statusLabel.Text = "连接成功: " + strings.TrimSpace(output)
			}
			statusLabel.Refresh()
			testBtn.Enable()
		}()
	}

	// 创建UI组件
	questionInput := widget.NewMultiLineEntry()
	questionInput.TextStyle = fyne.TextStyle{
		Bold:      true,
		Italic:    false,
		Monospace: false,
	}
	questionInput.SetPlaceHolder("在这里输入你的问题...")
	questionInput.Wrapping = fyne.TextWrapWord
	questionInput.Resize(fyne.NewSize(630, 120))

	// 使用Text组件显示结果
	resultEntry := widget.NewMultiLineEntry()
	resultEntry.SetText("Gemini 的回答将显示在这里...")
	resultEntry.Wrapping = fyne.TextWrapWord
	resultEntry.Disable() // 只读

	// 创建诊断按钮
	diagnosticBtn := widget.NewButton("环境诊断", nil)
	diagnosticBtn.OnTapped = func() {
		diagnosticBtn.Disable()
		fyne.CurrentApp().Queue().Submit(func() {
			resultEntry.SetText("⏳ 正在诊断服务器环境...")
			resultEntry.Refresh()
		})

		go func() {
			output, err := runDiagnostic()
			fyne.CurrentApp().Queue().Submit(func() {
				diagnosticBtn.Enable()
				if err != nil {
					resultEntry.SetText(fmt.Sprintf("❌ 诊断失败:\n%v\n\n输出:\n%s", err, output))
				} else {
					resultEntry.SetText("🔍 环境诊断结果:\n\n" + output)
				}
				resultEntry.Refresh()
			})
		}()
	}

	// 创建一个滚动容器来包装结果文本
	resultScroll := container.NewVScroll(container.NewPadded(resultEntry))
	resultScroll.SetMinSize(fyne.NewSize(630, 250)) // 稍微减小高度以适应新增的状态栏

	submitBtn := widget.NewButton("发送问题", nil)
	submitBtn.Importance = widget.HighImportance

	// 创建顶部工具栏
	toolbar := container.NewHBox(
		testBtn,
		widget.NewLabel("  |  "), // 分隔符
		diagnosticBtn,
		widget.NewLabel("  |  "), // 分隔符
		statusLabel,
	)

	// 使用 VBox 布局
	content := container.NewVBox(
		container.NewPadded(toolbar),
		container.NewPadded(questionInput),
		container.NewPadded(resultScroll),
		container.NewHBox(
			widget.NewLabel(""), // 左边距
			submitBtn,
			widget.NewLabel(""), // 右边距
		),
	)

	// 包装在滚动容器中
	mainScroll := container.NewScroll(content)
	myWindow.SetContent(mainScroll)

	// 定义按钮事件处理
	submitBtn.OnTapped = func() {
		question := questionInput.Text
		if strings.TrimSpace(question) == "" {
			fyne.CurrentApp().Queue().Submit(func() {
				resultEntry.SetText("错误：问题不能为空！")
				resultEntry.Refresh()
			})
			return
		}

		submitBtn.Disable()
		questionInput.Disable()
		fyne.CurrentApp().Queue().Submit(func() {
			resultEntry.SetText("⏳ 正在向Gemini提问，请稍候...")
			resultEntry.Refresh()
		})

		go func() {
			output, err := askGemini("gemini", question)
			fyne.CurrentApp().Queue().Submit(func() {
				submitBtn.Enable()
				questionInput.Enable()
				if err != nil {
					resultEntry.SetText(fmt.Sprintf("❌ 执行出错:\n%v\n\n服务器返回信息:\n%s", err, output))
				} else {
					resultEntry.SetText(output)
				}
				resultEntry.Refresh()
			})
		}()
	}

	// 添加快捷键支持
	if _, ok := myApp.(desktop.App); ok {
		myWindow.Canvas().AddShortcut(&desktop.CustomShortcut{
			KeyName:  fyne.KeyReturn,
			Modifier: desktop.ControlModifier,
		}, func(shortcut fyne.Shortcut) {
			if !submitBtn.Disabled() {
				submitBtn.OnTapped()
			}
		})
	}

	myWindow.ShowAndRun()
}

// askGemini 封装了SSH连接和命令执行的逻辑
func askGemini(executable, prompt string) (string, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("无法读取私钥文件 '%s': %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("无法解析私钥: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	serverAddr := host + ":" + port
	client, err := ssh.Dial("tcp", serverAddr, config)
	if err != nil {
		return "", fmt.Errorf("无法连接到SSH服务器 '%s': %w", serverAddr, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("无法创建SSH会话: %w", err)
	}
	defer session.Close()

	// 构建完整的命令，包含环境加载和gemini执行
	command := fmt.Sprintf(`
# 加载NVM环境
export NVM_DIR="$HOME/.nvm"
if [ -s "$NVM_DIR/nvm.sh" ]; then
    . "$NVM_DIR/nvm.sh"
fi

# 使用nvm加载Node.js
nvm use default 2>/dev/null || nvm use node 2>/dev/null || echo "使用系统Node.js"

# 执行gemini命令
echo "正在执行gemini命令..."
gemini "%s"
`, strings.ReplaceAll(prompt, `"`, `\"`))

	// 优先用 /bin/bash，如果失败再用 sh
	cmd := fmt.Sprintf("/bin/bash -c %s", strconv.Quote(command))
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		// fallback to sh
		cmd2 := fmt.Sprintf("sh -c %s", strconv.Quote(command))
		output2, err2 := session.CombinedOutput(cmd2)
		if err2 != nil {
			return string(output) + "\n\n" + string(output2), fmt.Errorf("gemini命令执行失败: %w (sh也失败: %w)", err, err2)
		}
		return string(output2), nil
	}
	return string(output), nil
}
