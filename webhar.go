package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 定义HAR文件结构体
type HAR struct {
	Log Log `json:"log"`
}

type Log struct {
	Version string  `json:"version"`
	Creator Creator `json:"creator"`
	Pages   []Page  `json:"pages,omitempty"`
	Entries []Entry `json:"entries"`
}

type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Page struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	StartTime string `json:"startedDateTime"`
}

type Entry struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Time     float64  `json:"time"`
}

type Request struct {
	Method  string   `json:"method"`
	URL     string   `json:"url"`
	Headers []Header `json:"headers"`
}

type Response struct {
	Status     int      `json:"status"`
	StatusText string   `json:"statusText"`
	Headers    []Header `json:"headers"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// 全局变量，用于存储当前解析的HAR数据
var currentHARData *HAR

// 从URL中提取域名
func extractDomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}

// 提取所有唯一域名
func extractUniqueDomains(harData *HAR) []string {
	domainMap := make(map[string]bool)
	for _, entry := range harData.Log.Entries {
		domain := extractDomain(entry.Request.URL)
		if domain != "" {
			domainMap[domain] = true
		}
	}

	var domains []string
	for domain := range domainMap {
		domains = append(domains, domain)
	}
	return domains
}

// 格式化文件大小
func formatFileSize(size int) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d byte", size)
	}
}

// 生成CSV内容（GBK编码）
func generateCSV(domains []string) []byte {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入标题行
	writer.Write([]string{"域名"})

	// 写入域名数据
	for _, domain := range domains {
		writer.Write([]string{domain})
	}

	writer.Flush()

	// 将UTF-8转换为GBK
	encoder := simplifiedchinese.GBK.NewEncoder()
	gbkBuf, _, _ := transform.Bytes(encoder, buf.Bytes())

	return gbkBuf
}

var htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>HAR Viewer</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
        }
        h1 {
            color: #333;
        }
        .file-upload {
            margin-bottom: 20px;
        }
        .har-info {
            background-color: #f0f0f0;
            padding: 10px;
            margin-bottom: 20px;
            border-radius: 5px;
        }
        .entries-list {
            list-style-type: none;
            padding: 0;
        }
        .entry-item {
            background-color: #f9f9f9;
            padding: 10px;
            margin-bottom: 10px;
            border-radius: 5px;
            cursor: pointer;
        }
        .entry-item:hover {
            background-color: #e9e9e9;
        }
        .entry-detail {
            background-color: #e0e0e0;
            padding: 15px;
            margin-top: 10px;
            border-radius: 5px;
            display: none;
        }
        .entry-detail.show {
            display: block;
        }
        .request-method {
            font-weight: bold;
            margin-right: 10px;
        }
        .status-code {
            margin-left: 10px;
            font-weight: bold;
        }
        .status-200 {
            color: green;
        }
        .status-300 {
            color: orange;
        }
        .status-400 {
            color: red;
        }
        .status-500 {
            color: darkred;
        }
        /* 统一按钮样式 */
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            margin: 5px;
            font-size: 14px;
            transition: background-color 0.3s ease;
        }
        
        /* 下载按钮样式 */
        .download-btn {
            background-color: #9E9E9E;
            color: white;
            margin-top: 10px;
        }
        .download-btn:hover {
            background-color: #757575;
        }
        
        /* 上传按钮样式 */
        .upload-btn {
            background-color: #2196F3;
            color: white;
        }
        .upload-btn:hover {
            background-color: #0b7dda;
        }
        
        /* 重新加载按钮样式 */
        .reload-btn {
            background-color: #f44336;
            color: white;
        }
        .reload-btn:hover {
            background-color: #da190b;
        }
        
        /* 文件选择按钮样式 */
        .file-input {
            padding: 8px 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            background-color: #f9f9f9;
            cursor: pointer;
            font-size: 14px;
            transition: all 0.3s ease;
            margin: 5px;
        }
        .file-input:hover {
            border-color: #2196F3;
            background-color: #f0f7ff;
        }
        
        /* 请求数量美化样式 */
        .method-count {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 12px;
            font-weight: bold;
            cursor: pointer;
            margin: 0 5px;
            transition: all 0.3s ease;
        }
        
        /* GET请求数量样式 */
        .method-count.get {
            background-color: #2196F3;
            color: white;
        }
        .method-count.get:hover {
            background-color: #0b7dda;
            transform: scale(1.05);
        }
        
        /* POST请求数量样式 */
        .method-count.post {
            background-color: #4CAF50;
            color: white;
        }
        .method-count.post:hover {
            background-color: #45a049;
            transform: scale(1.05);
        }
        
        /* 其他请求数量样式 */
        .method-count.other {
            background-color: #ff9800;
            color: white;
        }
        .method-count.other:hover {
            background-color: #e68a00;
            transform: scale(1.05);
        }
        
        /* 加载遮罩层样式 */
        #loading-mask {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0, 0, 0, 0.5);
            z-index: 9999;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        
        .loading-content {
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            text-align: center;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        }
        
        /* 圆环加载动画 */
        .loading-spinner {
            width: 50px;
            height: 50px;
            border: 5px solid #f3f3f3;
            border-top: 5px solid #2196F3;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin: 0 auto 15px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        /* 错误提示弹窗样式 */
        #error-modal {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0, 0, 0, 0.5);
            z-index: 10000;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        
        .modal-content {
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
            text-align: center;
            min-width: 300px;
        }
        
        .error-message {
            color: #f44336;
            font-size: 16px;
            font-weight: bold;
            margin: 0;
        }
        
        .progress-container {
            width: 100%;
            background-color: #f0f0f0;
            border-radius: 5px;
            margin: 5px 0;
            height: 10px;
        }
        .progress-bar {
            height: 100%;
            background-color: #4CAF50;
            border-radius: 5px;
            width: 0%;
        }
        .time-text {
            font-size: 12px;
            color: #666;
        }
        .entries-table {
            width: 94%;
            margin: 0 3%;
            border-collapse: collapse;
            table-layout: fixed;
        }
        .entries-table th,
        .entries-table td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
            overflow: hidden;
        }
        .entries-table th.url-col,
        .entries-table td.url-col {
            max-width: 70%;
            width: 70%;
        }
        .url-text {
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            display: inline-block;
            max-width: 100%;
        }
        .entry-detail {
            word-wrap: break-word;
            word-break: break-all;
        }
        .entry-detail div {
            word-wrap: break-word;
            word-break: break-all;
        }
        .entry-detail ul {
            padding-left: 20px;
        }
        .entry-detail li {
            margin: 5px 0;
            word-wrap: break-word;
            word-break: break-all;
        }
        .sort-indicator {
            cursor: pointer;
            margin-left: 5px;
            font-size: 12px;
        }
        .entries-table th {
            cursor: pointer;
        }
        .entries-table th {
            background-color: #f2f2f2;
            font-weight: bold;
        }
        .entries-table tr:hover {
            background-color: #f5f5f5;
        }
        .entries-table tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        .table-container {
            overflow-x: auto;
            margin: 15px 0;
        }
        
        /* 响应式布局 */
        @media (max-width: 768px) {
            body {
                margin: 10px;
            }
            
            h1, h2 {
                font-size: 1.5em;
            }
            
            .file-upload form {
                flex-direction: column;
                align-items: flex-start;
            }
            
            .btn {
                width: 100%;
                margin: 5px 0;
            }
            
            .file-input {
                width: 100%;
                margin: 5px 0;
            }
            
            .har-info {
                padding: 10px;
            }
            
            .entries-table {
                width: 100%;
                margin: 0;
            }
            
            .entries-table th,
            .entries-table td {
                padding: 5px;
                font-size: 12px;
            }
            
            .url-text {
                font-size: 12px;
            }
        }
        
        @media (max-width: 480px) {
            h1, h2 {
                font-size: 1.2em;
            }
            
            .entries-table th,
            .entries-table td {
                padding: 3px;
                font-size: 10px;
            }
        }
    </style>
</head>
<body>
    <h1>HAR Viewer</h1>
    
    <div class="file-upload" style="margin: 20px 0;">
        <form action="/upload" method="post" enctype="multipart/form-data" style="display: flex; flex-wrap: wrap; align-items: center;" onsubmit="showLoadingMask()">
            <input type="file" name="harfile" accept=".har" class="file-input" style="margin: 5px;">
            <input type="submit" value="上传HAR文件" class="btn upload-btn">
            <button type="button" onclick="location.href='/reload'" class="btn reload-btn">重新加载</button>
        </form>
    </div>
    
    {{if .HARData}}
    <div class="har-info">
        <h2>HAR文件信息</h2>
        <p>文件名: {{.FileName}}</p>
        <p>文件大小: {{.FileSize}} bytes</p>
        <p>请求数量: {{.MethodCountText}}</p>
        <a href="/download-csv" class="btn download-btn">下载域名CSV文件</a>
    </div>
    
    <h2>请求列表</h2>
    <div class="table-container">
        <table class="entries-table" id="entries-table">
            <thead>
                <tr>
                    <th class="method-col" onclick="sortEntries('method')">
                        方法 <span class="sort-indicator" id="sort-method">↕</span>
                    </th>
                    <th class="url-col" onclick="sortEntries('url')">
                        URL <span class="sort-indicator" id="sort-url">↕</span>
                    </th>
                    <th class="time-col" onclick="sortEntries('time')">
                        耗时 <span class="sort-indicator" id="sort-time">↕</span>
                    </th>
                </tr>
            </thead>
            <tbody id="entries-list">
                {{range $i, $entry := .HARData.Log.Entries}}
                <tr class="entry-item" onclick="toggleDetail(this)" data-method="{{$entry.Request.Method}}" data-url="{{$entry.Request.URL}}" data-time="{{$entry.Time}}">
                    <td class="method-col">
                        <span class="request-method">{{$entry.Request.Method}}</span>
                    </td>
                    <td class="url-col">
                        <div>
                            <span class="url-text">{{$entry.Request.URL}}</span>
                            <span class="status-code status-{{$entry.Response.Status}}">{{$entry.Response.Status}} {{$entry.Response.StatusText}}</span>
                        </div>
                        <div class="progress-container">
                            <div class="progress-bar" style="width: {{$entry.Time}}%;"></div>
                        </div>
                    </td>
                    <td class="time-col">
                        <span class="time-text">{{printf "%.2f" $entry.Time}} ms</span>
                    </td>
                </tr>
                <tr class="entry-detail" style="display: none;">
                    <td colspan="3">
                        <div style="padding: 15px; background-color: #e0e0e0; border-radius: 5px;">
                            <h3>请求详情</h3>
                            <p><strong>URL:</strong> {{$entry.Request.URL}}</p>
                            <p><strong>方法:</strong> {{$entry.Request.Method}}</p>
                            <p><strong>状态:</strong> {{$entry.Response.Status}} {{$entry.Response.StatusText}}</p>
                            <p><strong>耗时:</strong> {{printf "%.2f" $entry.Time}} ms</p>
                            
                            <h4>请求头</h4>
                            <ul>
                                {{range $header := $entry.Request.Headers}}
                                <li>{{$header.Name}}: {{$header.Value}}</li>
                                {{end}}
                            </ul>
                            
                            <h4>响应头</h4>
                            <ul>
                                {{range $header := $entry.Response.Headers}}
                                <li>{{$header.Name}}: {{$header.Value}}</li>
                                {{end}}
                            </ul>
                        </div>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
    
    {{end}}
    
    <!-- 加载遮罩层 -->
    <div id="loading-mask" style="display: none;">
        <div class="loading-content">
            <div class="loading-spinner"></div>
            <p>文件正在读取，请稍后...</p>
        </div>
    </div>
    
    <!-- 错误提示弹窗 -->
    <div id="error-modal" style="display: none;">
        <div class="modal-content">
            <p class="error-message">上传HAR文件失败，请检查文件格式是否正确！</p>
        </div>
    </div>
    
    <script>
        // 显示加载遮罩层
        function showLoadingMask() {
            // 300ms后显示遮罩层，避免快速上传时闪烁
            setTimeout(function() {
                const mask = document.getElementById('loading-mask');
                if (mask) {
                    mask.style.display = 'flex';
                }
            }, 300);
        }
        
        // 检查URL参数，显示错误弹窗
        function checkError() {
            const urlParams = new URLSearchParams(window.location.search);
            if (urlParams.get('error') === '1') {
                const modal = document.getElementById('error-modal');
                if (modal) {
                    modal.style.display = 'flex';
                    
                    // 5秒后自动隐藏弹窗
                    const timeoutId = setTimeout(function() {
                        modal.style.display = 'none';
                        // 移除URL中的error参数，避免刷新页面后再次显示弹窗
                        window.history.replaceState({}, document.title, window.location.pathname);
                    }, 5000);
                    
                    // 点击页面任一位置关闭弹窗
                    function closeModal() {
                        modal.style.display = 'none';
                        clearTimeout(timeoutId);
                        // 移除URL中的error参数，避免刷新页面后再次显示弹窗
                        window.history.replaceState({}, document.title, window.location.pathname);
                        // 移除事件监听器，避免内存泄漏
                        document.removeEventListener('click', closeModal);
                    }
                    
                    // 添加点击事件监听器
                    document.addEventListener('click', closeModal);
                }
            }
        }
        
        // 页面加载时检查错误
        window.addEventListener('load', checkError);
        
        // 切换详情显示
        function toggleDetail(row) {
            const detail = row.nextElementSibling;
            if (detail && detail.classList.contains('entry-detail')) {
                if (detail.style.display === 'none') {
                    detail.style.display = 'table-row';
                } else {
                    detail.style.display = 'none';
                }
            }
        }
        
        // 获取所有数据行和对应的详情行
        function getRowPairs() {
            const list = document.getElementById('entries-list');
            if (!list) return [];
            
            const rowPairs = [];
            let current = list.firstElementChild;
            while (current) {
                if (current.classList.contains('entry-item')) {
                    var dataRow = current;
                    var detailRow = current.nextElementSibling;
                    if (detailRow && detailRow.classList.contains('entry-detail')) {
                        rowPairs.push({data: dataRow, detail: detailRow});
                    }
                    current = detailRow.nextElementSibling;
                } else {
                    current = current.nextElementSibling;
                }
            }
            return rowPairs;
        }
        
        // 排序方向状态管理
        let sortDirections = {
            method: 'asc',
            url: 'asc',
            time: 'asc'
        };
        
        // 排序功能实现
        function sortEntries(sortBy) {
            const list = document.getElementById('entries-list');
            if (!list) return;
            
            // 获取所有数据行和对应的详情行
            const rowPairs = getRowPairs();
            
            const direction = sortDirections[sortBy];
            
            // 对数据行进行排序
            for (let i = 0; i < rowPairs.length; i++) {
                for (let j = i + 1; j < rowPairs.length; j++) {
                    let a = rowPairs[i].data;
                    let b = rowPairs[j].data;
                    let aVal, bVal, result = 0;
                    
                    if (sortBy === 'method') {
                        aVal = a.dataset.method;
                        bVal = b.dataset.method;
                        result = aVal.localeCompare(bVal);
                    } else if (sortBy === 'url') {
                        aVal = a.dataset.url;
                        bVal = b.dataset.url;
                        result = aVal.localeCompare(bVal);
                    } else if (sortBy === 'time') {
                        aVal = parseFloat(a.dataset.time);
                        bVal = parseFloat(b.dataset.time);
                        result = aVal - bVal;
                    }
                    
                    // 根据排序方向调整结果
                    if (direction === 'desc') {
                        result = -result;
                    }
                    
                    // 交换位置
                    if (result > 0) {
                        let temp = rowPairs[i];
                        rowPairs[i] = rowPairs[j];
                        rowPairs[j] = temp;
                    }
                }
            }
            
            // 切换排序方向
            sortDirections[sortBy] = direction === 'asc' ? 'desc' : 'asc';
            
            // 更新表头排序指示器
            updateSortIndicators();
            
            // 清空列表并重新添加排序后的项（包括对应的详情行）
            list.innerHTML = '';
            for (let i = 0; i < rowPairs.length; i++) {
                list.appendChild(rowPairs[i].data);
                list.appendChild(rowPairs[i].detail);
            }
        }
        
        // 更新排序指示器
        function updateSortIndicators() {
            // 重置所有指示器为默认状态
            document.getElementById('sort-method').textContent = '↕';
            document.getElementById('sort-url').textContent = '↕';
            document.getElementById('sort-time').textContent = '↕';
            
            // 对于每个排序字段，更新对应的指示器
            var fields = ['method', 'url', 'time'];
            for (var i = 0; i < fields.length; i++) {
                var field = fields[i];
                var direction = sortDirections[field];
                var indicator = document.getElementById('sort-' + field);
                if (indicator) {
                    if (direction === 'asc') {
                        indicator.textContent = '↑';
                    } else {
                        indicator.textContent = '↓';
                    }
                }
            }
        }
        
        // 按请求方法排序，将指定方法的请求置顶
        function sortByMethod(method) {
            const list = document.getElementById('entries-list');
            if (!list) return;
            
            // 获取所有数据行和对应的详情行
            const rowPairs = getRowPairs();
            
            // 按方法类型排序，将指定方法的请求置顶
            rowPairs.sort(function(a, b) {
                var aMethod = a.data.dataset.method;
                var bMethod = b.data.dataset.method;
                
                // 检查a是否匹配指定方法
                var aMatch = (method === 'OTHER' && aMethod !== 'GET' && aMethod !== 'POST') || aMethod === method;
                // 检查b是否匹配指定方法
                var bMatch = (method === 'OTHER' && bMethod !== 'GET' && bMethod !== 'POST') || bMethod === method;
                
                // 如果a匹配而b不匹配，a排在前面
                if (aMatch && !bMatch) {
                    return -1;
                }
                // 如果b匹配而a不匹配，b排在前面
                if (!aMatch && bMatch) {
                    return 1;
                }
                // 如果都匹配或都不匹配，保持原顺序
                return 0;
            });
            
            // 清空列表并重新添加排序后的项（包括对应的详情行）
            list.innerHTML = '';
            for (var i = 0; i < rowPairs.length; i++) {
                list.appendChild(rowPairs[i].data);
                list.appendChild(rowPairs[i].detail);
            }
        }
        
        // 初始化进度条
        document.addEventListener('DOMContentLoaded', function() {
            const progressBars = document.querySelectorAll('.progress-bar');
            if (progressBars.length === 0) return;
            
            let maxTime = 0;
            
            // 找到最大耗时
            progressBars.forEach(bar => {
                // 在table结构中，time-text位于同一行的第三个td中
                const row = bar.closest('tr');
                if (row) {
                    const timeCell = row.querySelector('.time-col');
                    if (timeCell) {
                        const timeText = timeCell.querySelector('.time-text');
                        if (timeText) {
                            const time = parseFloat(timeText.textContent);
                            if (time > maxTime) {
                                maxTime = time;
                            }
                        }
                    }
                }
            });
            
            // 设置进度条宽度
            progressBars.forEach(bar => {
                const row = bar.closest('tr');
                if (row && maxTime > 0) {
                    const timeCell = row.querySelector('.time-col');
                    if (timeCell) {
                        const timeText = timeCell.querySelector('.time-text');
                        if (timeText) {
                            const time = parseFloat(timeText.textContent);
                            const width = (time / maxTime) * 100;
                            bar.style.width = width + '%';
                        }
                    }
                }
            });
        });
    </script>
</body>
</html>`

// 设置路由
func setupRoutes() {
	// 定义路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download-csv", downloadCSVHandler)
	http.HandleFunc("/reload", reloadHandler)
}

// 重新加载处理函数
func reloadHandler(w http.ResponseWriter, r *http.Request) {
	// 清空全局HAR数据
	currentHARData = nil
	// 重定向到首页
	http.Redirect(w, r, "/", http.StatusFound)
}

// 下载CSV处理函数
func downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
	if currentHARData == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// 提取唯一域名
	domains := extractUniqueDomains(currentHARData)

	// 生成CSV内容（GBK编码）
	csvContent := generateCSV(domains)

	// 设置响应头
	w.Header().Set("Content-Type", "text/csv; charset=GBK")
	w.Header().Set("Content-Disposition", "attachment; filename=domains.csv")

	// 写入响应
	w.Write(csvContent)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("har").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析模板失败: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, map[string]interface{}{
		"HARData":         nil,
		"MethodCountText": template.HTML(""),
	})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// 解析表单
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, fmt.Sprintf("解析表单失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("harfile")
	if err != nil {
		http.Error(w, fmt.Sprintf("获取文件失败: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 读取文件内容
	content, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取文件失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 解析HAR文件
	var harData HAR
	err = json.Unmarshal(content, &harData)
	if err != nil {
		// 解析失败，重载页面到初始状态并显示错误
		http.Redirect(w, r, "/?error=1", http.StatusFound)
		return
	}

	// 统计请求方法数量
	getCount := 0
	postCount := 0
	otherCount := 0
	for _, entry := range harData.Log.Entries {
		switch entry.Request.Method {
		case "GET":
			getCount++
		case "POST":
			postCount++
		default:
			otherCount++
		}
	}

	// 生成请求数量显示文本
	var methodCounts []string
	if getCount > 0 {
		methodCounts = append(methodCounts, fmt.Sprintf("<span class=\"method-count get\" onclick=\"sortByMethod('GET')\">GET %d个</span>", getCount))
	}
	if postCount > 0 {
		methodCounts = append(methodCounts, fmt.Sprintf("<span class=\"method-count post\" onclick=\"sortByMethod('POST')\">POST %d个</span>", postCount))
	}
	if otherCount > 0 {
		methodCounts = append(methodCounts, fmt.Sprintf("<span class=\"method-count other\" onclick=\"sortByMethod('OTHER')\">其他 %d个</span>", otherCount))
	}
	methodCountText := strings.Join(methodCounts, "    ")

	// 存储到全局变量
	currentHARData = &harData

	// 渲染模板
	tmpl, err := template.New("har").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析模板失败: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, map[string]interface{}{
		"HARData":         harData,
		"FileName":        header.Filename,
		"FileSize":        formatFileSize(len(content)),
		"MethodCountText": template.HTML(methodCountText),
	})
}
