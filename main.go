package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/w7panel/w7panel-kubeblocks/pkg/helm"
	"github.com/w7panel/w7panel-kubeblocks/pkg/kbcli"
)

var (
	k8sClient *kubernetes.Clientset
)

// LoggingMiddleware 打印请求信息的中间件
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		log.Printf("=== Request ===")
		log.Printf("Method: %s", method)
		log.Printf("Path: %s", path)

		// 打印 Query 参数
		if len(c.Request.URL.Query()) > 0 {
			log.Printf("Query Parameters:")
			for key, values := range c.Request.URL.Query() {
				for _, value := range values {
					log.Printf("  %s: %s", key, value)
				}
			}
		} else {
			log.Printf("Query Parameters: none")
		}

		// 打印 Header
		if len(c.Request.Header) > 0 {
			log.Printf("Headers:")
			for key, values := range c.Request.Header {
				for _, value := range values {
					log.Printf("  %s: %s", key, value)
				}
			}
		} else {
			log.Printf("Headers: none")
		}

		log.Printf("===============")

		c.Next()
	}
}

// AuthMiddleware API Key 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := os.Getenv("API_KEY")

		// 如果环境变量 API_KEY 没有设置，放行
		if apiKey == "" {
			c.Next()
			return
		}

		// 从查询参数或请求头获取 api-key
		clientAPIKey := c.Query("api-key")
		if clientAPIKey == "" {
			clientAPIKey = c.GetHeader("api-key")
		}

		// 验证 api-key 是否匹配
		if clientAPIKey != apiKey {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: Invalid or missing api-key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func initK8sClient() error {
	var config *rest.Config
	var err error

	// 尝试从集群内配置加载
	config, err = rest.InClusterConfig()
	if err != nil {
		// 如果在集群外运行,从 kubeconfig 文件加载
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return err
		}
	}

	k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

type AddonInfo struct {
	AddonName string   `json:"addon-name"`
	Files     []string `json:"files"`
}

func getAddons(c *gin.Context) {
	kodata := os.Getenv("KO_DATA_PATH")
	addonsDir := kodata + "/block-index/addons"

	var addons []AddonInfo

	entries, err := os.ReadDir(addonsDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		addonName := entry.Name()
		addonPath := filepath.Join(addonsDir, addonName)

		var yamlFiles []string
		hasVersionFile := false
		files, err := os.ReadDir(addonPath)
		if err != nil {
			log.Printf("Failed to read directory %s: %v", addonPath, err)
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if file.Name() == "1.0.1.yaml" {
				hasVersionFile = true
			}
			if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
				yamlFiles = append(yamlFiles, file.Name())
			}
		}

		if hasVersionFile {
			sort.Strings(yamlFiles)
			addons = append(addons, AddonInfo{
				AddonName: addonName,
				Files:     yamlFiles,
			})
		}
	}

	sort.Slice(addons, func(i, j int) bool {
		return addons[i].AddonName < addons[j].AddonName
	})

	c.JSON(http.StatusOK, gin.H{
		"addons": addons,
		"count":  len(addons),
	})
}

func getAddonFile(c *gin.Context) {
	addonName := c.Param("addon")
	filename := c.Param("file")

	// 安全检查:防止路径遍历攻击
	if strings.Contains(addonName, "..") || strings.Contains(filename, "..") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid path",
		})
		return
	}
	kodata := os.Getenv("KO_DATA_PATH")
	// 构建文件路径
	filePath := filepath.Join(kodata+"/block-index/addons", addonName, filename)

	// 检查文件扩展名
	if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only .yaml and .yml files are supported",
		})
		return
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 确保是文件而不是目录
	if fileInfo.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Not a file",
		})
		return
	}

	// 读取文件内容
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Header("Content-Type", "text/yaml; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

func getClusterSchema(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	clusterName := name + "-cluster"

	settings := cli.New()
	// 安全检查:防止路径遍历攻击
	chart, err := helm.LocateChart(settings, clusterName, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.String(http.StatusOK, string(chart.Schema))
}

func getConfigParams(c *gin.Context) {
	clusterName := c.DefaultQuery("cluster", "mysql57")
	namespace := c.DefaultQuery("namespace", "default")
	components := c.QueryArray("components")
	if len(components) == 0 {
		if component := strings.TrimSpace(c.Query("component")); component != "" {
			components = []string{component}
		}
	}

	wrapper, err := kbcli.New(clusterName, namespace, components...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      err.Error(),
			"cluster":    clusterName,
			"namespace":  namespace,
			"components": components,
		})
		return
	}

	data, err := wrapper.ToJson()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      err.Error(),
			"cluster":    clusterName,
			"namespace":  namespace,
			"components": components,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

func getConnectInfo(c *gin.Context) {
	clusterName := strings.TrimSpace(c.Query("cluster"))
	namespace := c.DefaultQuery("namespace", "default")
	componentName := strings.TrimSpace(c.Query("component"))
	instanceName := strings.TrimSpace(c.Query("instance"))
	clientType := strings.TrimSpace(c.Query("client"))

	wrapper, err := kbcli.NewConnectWrapper(namespace, clusterName, componentName, instanceName, clientType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"cluster":    clusterName,
			"namespace":  namespace,
			"component":  componentName,
			"instance":   instanceName,
			"clientType": clientType,
		})
		return
	}

	data, err := wrapper.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      err.Error(),
			"cluster":    clusterName,
			"namespace":  namespace,
			"component":  componentName,
			"instance":   instanceName,
			"clientType": clientType,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

type GetClusterManifestRequest struct {
	ReleaseName  string                 `json:"release_name" binding:"required"`
	ChartRef     string                 `json:"chart_ref" binding:"required"`
	ChartVersion string                 `json:"chart_version" binding:"required"`
	Values       map[string]interface{} `json:"values"`
	Namespace    string                 `json:"namespace"`
}

func checkKubeBlocksRelease(c *gin.Context) {
	namespace := "kb-system"
	releaseName := "kubeblocks"

	exists, err := helm.CheckReleaseExists(log.Default(), namespace, releaseName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check kubeblocks release: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exists":    exists,
		"namespace": namespace,
		"release":   releaseName,
	})
}

func getClusterManifest(c *gin.Context) {
	var req GetClusterManifestRequest

	// 绑定 JSON 请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// 设置 namespace
	settings := cli.New()
	if req.Namespace != "" {
		settings.SetNamespace(req.Namespace)
	}

	// 调用 TryRunInstallGetClusterManifest
	info, err := helm.TryRunInstallGetClusterManifest(
		context.Background(),
		log.Default(),
		settings,
		req.ReleaseName,
		req.ChartRef,
		req.ChartVersion,
		req.Values,
	)
	if err != nil {
		log.Printf("Failed to get cluster manifest: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get cluster manifest: " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, info.Object)
}

type InstallResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func installKubeBlocks(c *gin.Context) {
	kodata := os.Getenv("KO_DATA_PATH")
	if kodata == "" {
		c.JSON(http.StatusBadRequest, InstallResponse{
			Status:  "error",
			Message: "KO_DATA_PATH environment variable is not set",
		})
		return
	}

	// 路径安全检查: 防止路径遍历攻击
	if strings.Contains(kodata, "..") {
		c.JSON(http.StatusBadRequest, InstallResponse{
			Status:  "error",
			Message: "Invalid KO_DATA_PATH: path traversal detected",
		})
		return
	}

	// 确保 kodata 路径存在
	if _, err := os.Stat(kodata); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, InstallResponse{
			Status:  "error",
			Message: "KO_DATA_PATH directory does not exist",
		})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建 channel 用于发送日志
	logChan := make(chan string, 100)

	// 添加超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 启动 goroutine 执行安装脚本
	go func() {
		defer close(logChan)

		// 检查脚本是否存在
		scriptPath := filepath.Join(kodata, "install-kubeblocks.sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			logChan <- fmt.Sprintf("Error: Installation script not found at %s", scriptPath)
			logChan <- "status: error"
			return
		}

		// 执行安装脚本
		cmd := exec.CommandContext(ctx, "/bin/sh", scriptPath)
		cmd.Env = append(os.Environ(), fmt.Sprintf("KO_DATA_PATH=%s", kodata))

		logChan <- "Starting KubeBlocks installation..."
		logChan <- "---"

		// 流式输出脚本执行结果
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			logChan <- fmt.Sprintf("Error creating stdout pipe: %v", err)
			logChan <- "status: error"
			return
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			logChan <- fmt.Sprintf("Error creating stderr pipe: %v", err)
			logChan <- "status: error"
			return
		}

		if err := cmd.Start(); err != nil {
			logChan <- fmt.Sprintf("Error starting installation script: %v", err)
			logChan <- "status: error"
			return
		}

		// 使用 WaitGroup 确保 goroutine 正确退出
		var wg sync.WaitGroup

		// 读取 stdout
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				logChan <- scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				logChan <- fmt.Sprintf("Error reading stdout: %v", err)
			}
		}()

		// 读取 stderr
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				logChan <- scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				logChan <- fmt.Sprintf("Error reading stderr: %v", err)
			}
		}()

		// 等待命令完成
		err = cmd.Wait()
		wg.Wait() // 确保 goroutine 完成

		if err != nil {
			logChan <- fmt.Sprintf("Installation script failed: %v", err)
			logChan <- "status: error"
			return
		}

		logChan <- "status: completed"
	}()

	// 检测客户端断开连接
	notify := c.Request.Context().Done()

	// 通过 SSE 发送日志
	for {
		select {
		case logLine, ok := <-logChan:
			if !ok {
				// channel 已关闭，正常退出
				return
			}
			c.SSEvent("message", logLine)
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-notify:
			// 客户端断开连接，取消上下文停止后台任务
			cancel()
			return
		case <-ctx.Done():
			// 超时或其他上下文取消
			c.SSEvent("message", "Installation timed out or was cancelled")
			c.SSEvent("message", "status: error")
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
			return
		}
	}
}

// ShellCommandRequest 定义 shell 命令请求
type ShellCommandRequest struct {
	Name string `json:"name" binding:"required"`
}

// 执行 kbcli addon enable 命令
func enableAddon(c *gin.Context) {
	var req ShellCommandRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	log.Printf("Enabling addon: %s", req.Name)

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建 channel 用于发送日志
	logChan := make(chan string, 100)

	// 添加超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 启动 goroutine 执行命令
	go func() {
		defer close(logChan)

		// 执行 kbcli addon enable 命令
		cmd := exec.CommandContext(ctx, "kbcli", "addon", "enable", req.Name)

		logChan <- fmt.Sprintf("Starting addon enable: %s", req.Name)
		logChan <- "---"

		// 流式输出命令执行结果
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			logChan <- fmt.Sprintf("Error creating stdout pipe: %v", err)
			logChan <- "status: error"
			return
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			logChan <- fmt.Sprintf("Error creating stderr pipe: %v", err)
			logChan <- "status: error"
			return
		}

		if err := cmd.Start(); err != nil {
			logChan <- fmt.Sprintf("Error starting command: %v", err)
			logChan <- "status: error"
			return
		}

		// 使用 WaitGroup 确保 goroutine 正确退出
		var wg sync.WaitGroup

		// 读取 stdout
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				logChan <- scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				logChan <- fmt.Sprintf("Error reading stdout: %v", err)
			}
		}()

		// 读取 stderr
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				logChan <- scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				logChan <- fmt.Sprintf("Error reading stderr: %v", err)
			}
		}()

		// 等待命令完成
		err = cmd.Wait()
		wg.Wait()

		if err != nil {
			logChan <- fmt.Sprintf("Command failed: %v", err)
			logChan <- "status: error"
			return
		}

		logChan <- "status: completed"
	}()

	// 检测客户端断开连接
	notify := c.Request.Context().Done()

	// 通过 SSE 发送日志
	for {
		select {
		case logLine, ok := <-logChan:
			if !ok {
				return
			}
			c.SSEvent("message", logLine)
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-notify:
			cancel()
			return
		case <-ctx.Done():
			c.SSEvent("message", "Command timed out or was cancelled")
			c.SSEvent("message", "status: error")
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
			return
		}
	}
}

// 执行 kbcli addon disable 命令
func disableAddon(c *gin.Context) {
	var req ShellCommandRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	log.Printf("Disabling addon: %s", req.Name)

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建 channel 用于发送日志
	logChan := make(chan string, 100)

	// 添加超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 启动 goroutine 执行命令
	go func() {
		defer close(logChan)

		// 执行 kbcli addon disable 命令
		cmd := exec.CommandContext(ctx, "kbcli", "addon", "disable", req.Name)

		logChan <- fmt.Sprintf("Starting addon disable: %s", req.Name)
		logChan <- "---"

		// 流式输出命令执行结果
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			logChan <- fmt.Sprintf("Error creating stdout pipe: %v", err)
			logChan <- "status: error"
			return
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			logChan <- fmt.Sprintf("Error creating stderr pipe: %v", err)
			logChan <- "status: error"
			return
		}

		if err := cmd.Start(); err != nil {
			logChan <- fmt.Sprintf("Error starting command: %v", err)
			logChan <- "status: error"
			return
		}

		// 使用 WaitGroup 确保 goroutine 正确退出
		var wg sync.WaitGroup

		// 读取 stdout
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				logChan <- scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				logChan <- fmt.Sprintf("Error reading stdout: %v", err)
			}
		}()

		// 读取 stderr
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				logChan <- scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				logChan <- fmt.Sprintf("Error reading stderr: %v", err)
			}
		}()

		// 等待命令完成
		err = cmd.Wait()
		wg.Wait()

		if err != nil {
			logChan <- fmt.Sprintf("Command failed: %v", err)
			logChan <- "status: error"
			return
		}

		logChan <- "status: completed"
	}()

	// 检测客户端断开连接
	notify := c.Request.Context().Done()

	// 通过 SSE 发送日志
	for {
		select {
		case logLine, ok := <-logChan:
			if !ok {
				return
			}
			c.SSEvent("message", logLine)
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-notify:
			cancel()
			return
		case <-ctx.Done():
			c.SSEvent("message", "Command timed out or was cancelled")
			c.SSEvent("message", "status: error")
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
			return
		}
	}
}

func main() {

	log.Println("Kubernetes client initialized successfully")

	// 创建 Gin 路由
	router := gin.Default()

	// 添加全局日志中间件
	router.Use(LoggingMiddleware())

	// 健康检查（无需认证）
	router.GET("/health", healthCheck)

	// 创建需要认证的路由组
	api := router.Group("/api/v1")
	api.Use(AuthMiddleware())

	// KubeBlocks Release API
	api.GET("/kubeblocks/release/check", checkKubeBlocksRelease)

	// Addons API
	api.GET("/addons", getAddons)
	api.GET("/addons/:addon/:file", getAddonFile)
	api.POST("/addons/enable", enableAddon)
	api.POST("/addons/disable", disableAddon)

	// Cluster API
	api.GET("/cluster/:name/:version", getClusterSchema)
	api.POST("/cluster/manifest", getClusterManifest)
	api.GET("/config/params", getConfigParams)
	api.GET("/connect", getConnectInfo)

	// Installation API
	api.GET("/install/kubeblocks", installKubeBlocks)

	// 启动服务
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
