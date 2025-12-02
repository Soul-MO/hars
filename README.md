# HAR Viewer

HAR Viewer 是一个用于查看和分析 HAR (HTTP Archive) 文件的工具，提供 GUI 界面和 Web 界面

## 功能特性

### GUI 界面功能-服务端
- **端口号设置**：可自定义 Web 服务的端口，默认 8081
- **启动 Web 服务**：一键启动 Web 服务并自动打开浏览器
- **关闭 Web 服务**：安全关闭正在运行的 Web 服务
- **打开程序**：使用默认浏览器访问 Web 服务
- **退出程序**：关闭 Web 服务并退出 GUI 界面

### Web 界面功能-前端
- **上传 HAR 文件**：选择并上传 HAR 格式的文件
- **请求列表**：展示所有 HTTP 请求，支持点击查看详情
- **排序功能**：点击表头可按方法、URL 或耗时排序
- **下载域名 CSV**：提取所有唯一域名并保存为 CSV 文件
- **重新加载**：清空所有数据，重新开始

## 安装方法

### 直接运行
1. 从项目根目录下载 `harviewer.exe` 文件
2. 双击运行即可

### 源码编译
1. 确保已安装 Go 1.25.4 或更高版本
2. 克隆或下载项目源码
3. 进入项目目录，执行以下命令编译：
   ```bash
   go build -o harviewer.exe -ldflags="-H windowsgui -s -w" -trimpath
   ```

## 使用说明

1. **启动程序**：双击 `harviewer.exe` 文件
2. **设置端口**：在 GUI 界面输入自定义端口号（可选，默认 8081）
3. **启动服务**：点击「启动 web 服务」按钮，程序会自动打开浏览器
4. **上传文件**：在 Web 界面点击「选择文件」按钮，上传 HAR 文件
5. **查看数据**：在 Web 界面查看请求列表，点击请求可查看详情
6. **导出数据**：点击「下载域名 CSV」按钮，导出域名列表
7. **关闭服务**：在 GUI 界面点击「关闭 web 服务」按钮
8. **退出程序**：在 GUI 界面点击「退出程序」按钮

## 项目结构

```
hars/
├── README.md          # 项目说明文档
├── go.mod             # Go 模块依赖
├── go.sum             # 依赖校验文件
├── harviewer.exe      # 编译后的可执行文件
├── icon.ico           # 程序图标
├── icon.png           # PNG 格式图标
├── icon.rc            # 图标资源脚本
├── icon_windows_amd64.syso  # Windows 资源文件
├── main.go            # 主程序入口
├── versioninfo.json   # 版本信息配置
└── webhar.go          # Web 服务和 HAR 解析逻辑
```

## 技术栈

- **编程语言**：Go 1.25.4
- **GUI 框架**：Fyne v2.7.1
- **Web 框架**：Go 标准库 `net/http`
- **图标处理**：Windows 资源文件 (.rc, .syso)

## 编译说明

### 编译选项说明
- `-ldflags="-H windowsgui"`：生成 GUI 程序，不显示命令行窗口
- `-ldflags="-s -w"`：去除调试信息和符号表，减小文件体积
- `-trimpath`：去除编译路径信息
- `-gcflags="-m -l"`：优化编译，禁用内联

### 重新生成资源文件
如果修改了图标或版本信息，需要重新生成资源文件：

```bash
# 安装 rsrc 工具
go install github.com/akavel/rsrc@latest

# 生成资源文件
rsrc -manifest versioninfo.json -ico icon.ico -o icon_windows_amd64.syso
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题或建议，欢迎通过 GitHub Issues 反馈。
