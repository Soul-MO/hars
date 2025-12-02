package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	httpServer *http.Server
	port       = "8081"
)

func main() {
	// 设置路由
	setupRoutes()

	// 创建Fyne应用
	myApp := app.New()
	// 设置应用图标
	if icon, err := fyne.LoadResourceFromPath("icon.ico"); err == nil {
		myApp.SetIcon(icon)
	}
	myWindow := myApp.NewWindow("HAR Viewer")
	myWindow.Resize(fyne.NewSize(400, 200))

	// 端口号输入框
	portEntry := widget.NewEntry()
	portEntry.SetText(port)
	portEntry.SetPlaceHolder("请输入端口号")

	// 先声明stopBtn和openBtn变量，使用nil初始化
	var stopBtn *widget.Button
	var openBtn *widget.Button

	// 启动按钮
	startBtn := widget.NewButton("启动web服务", nil)
	startBtn.OnTapped = func() {
		port = portEntry.Text
		if port == "" {
			port = "8081"
		}

		// 启动HTTP服务器
		go func() {
			httpServer = &http.Server{
				Addr:    ":" + port,
				Handler: nil, // 使用默认的http.ServeMux
			}

			fmt.Printf("HAR Viewer 已启动，访问地址: http://localhost:%s\n", port)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("启动服务器失败: %v\n", err)
			}
		}()

		// 等待服务器启动，然后自动打开浏览器
		url := fmt.Sprintf("http://localhost:%s", port)
		openURL(url)

		startBtn.Disable()
		portEntry.Disable()
		stopBtn.Enable()
		openBtn.Enable() // 启用打开程序按钮
	}

	// 关闭web服务按钮
	stopBtn = widget.NewButton("关闭web服务", func() {
		if httpServer != nil {
			// 关闭HTTP服务器
			if err := httpServer.Close(); err != nil {
				fmt.Printf("关闭服务器失败: %v\n", err)
			} else {
				fmt.Printf("HAR Viewer 已关闭\n")
			}
			httpServer = nil
		}

		startBtn.Enable()
		portEntry.Enable()
		stopBtn.Disable()
		openBtn.Disable() // 禁用打开程序按钮
	})
	stopBtn.Disable() // 初始状态为禁用

	// 打开程序按钮
	openBtn = widget.NewButton("打开程序", func() {
		url := fmt.Sprintf("http://localhost:%s", port)
		openURL(url)
	})
	openBtn.Disable() // 初始状态为禁用

	// 退出程序按钮
	quitBtn := widget.NewButton("退出程序", func() {
		// 先关闭服务器，再退出程序
		if httpServer != nil {
			httpServer.Close()
		}
		myApp.Quit()
	})

	// 使用说明
	usageLabel := widget.NewLabel("📖 使用说明")
	usageLabel.TextStyle = fyne.TextStyle{Bold: true}

	guiIntroLabel := widget.NewLabel("1. GUI界面功能")
	guiIntroLabel.TextStyle = fyne.TextStyle{Bold: true}
	guiDetailLabel := widget.NewLabel("   • 端口号：设置Web服务的端口，默认8081\n   • 启动web服务：启动Web服务并自动打开浏览器\n   • 关闭web服务：关闭正在运行的Web服务\n   • 打开程序：使用默认浏览器访问Web服务\n   • 退出程序：关闭Web服务并退出GUI界面")

	htmlIntroLabel := widget.NewLabel("2. Web界面功能")
	htmlIntroLabel.TextStyle = fyne.TextStyle{Bold: true}
	htmlDetailLabel := widget.NewLabel("   • 上传HAR文件：选择并上传HAR格式的文件\n   • 请求列表：展示所有HTTP请求，支持点击查看详情\n   • 排序功能：点击表头可按方法、URL或耗时排序\n   • 下载域名CSV：提取所有唯一域名并保存为CSV文件\n   • 重新加载：清空所有数据，重新开始")

	// 布局设计
	content := container.NewVBox(
		widget.NewLabel("端口号:"),
		portEntry,
		container.NewHBox(
			startBtn,
			stopBtn,
			openBtn,
			quitBtn,
		),
		usageLabel,
		guiIntroLabel,
		guiDetailLabel,
		htmlIntroLabel,
		htmlDetailLabel,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

// 打开URL
func openURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		fmt.Printf("不支持的操作系统: %s\n", runtime.GOOS)
		return
	}

	if err := exec.Command(cmd, args...).Start(); err != nil {
		fmt.Printf("打开浏览器失败: %v\n", err)
	}
}
