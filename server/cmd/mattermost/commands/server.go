// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// server 子命令实现，负责启动 Mattermost 服务器
package commands

import (
	"bytes"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/v8/channels/api4"
	"github.com/mattermost/mattermost/server/v8/channels/app"
	"github.com/mattermost/mattermost/server/v8/channels/utils"
	"github.com/mattermost/mattermost/server/v8/channels/web"
	"github.com/mattermost/mattermost/server/v8/channels/wsapi"
	"github.com/mattermost/mattermost/server/v8/config"
)

// serverCmd 定义 "server" 子命令
var serverCmd = &cobra.Command{
	Use:          "server",
	Short:        "Run the Mattermost server",
	RunE:         serverCmdF,
	SilenceUsage: true,
}

func init() {
	// 将 server 命令注册为 RootCmd 的子命令
	RootCmd.AddCommand(serverCmd)
	// 如果直接运行 "mattermost" 而没有子命令，默认执行 server 命令
	RootCmd.RunE = serverCmdF
}

// serverCmdF 执行服务器启动流程
func serverCmdF(command *cobra.Command, args []string) error {
	// 创建中断信号通道
	interruptChan := make(chan os.Signal, 1)

	// 初始化国际化文件
	if err := utils.TranslationsPreInit(); err != nil {
		return errors.Wrap(err, "unable to load Mattermost translation files")
	}

	// 加载自定义配置默认值
	customDefaults, err := loadCustomDefaults()
	if err != nil {
		mlog.Warn("Error loading custom configuration defaults: " + err.Error())
	}

	// 从 DSN（数据源名称）创建配置存储
	configStore, err := config.NewStoreFromDSN(getConfigDSN(command, config.GetEnvironment()), false, customDefaults, true)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration")
	}
	defer configStore.Close()

	// 运行服务器
	return runServer(configStore, interruptChan)
}

// runServer 启动并运行 Mattermost 服务器，直到收到中断信号
func runServer(configStore *config.Store, interruptChan chan os.Signal) error {
	// 设置最高级别的 traceback，用于在崩溃时打印所有 goroutine 信息
	// 这对于调试和调查崩溃非常重要（参见 golang.org/issue/13161）
	debug.SetTraceback("crash")

	// 配置服务器启动选项
	// 选项的顺序很重要，app.Config 选项需要读取 app.StartMetrics 选项
	options := []app.Option{
		app.StartMetrics,        // 启动指标收集服务
		app.ConfigStore(configStore), // 传入配置存储
		app.RunEssentialJobs,    // 运行必要的后台任务
		app.JoinCluster,         // 加入集群（如果启用了集群）
	}
	
	// 创建服务器实例
	server, err := app.NewServer(options...)
	if err != nil {
		mlog.Error(err.Error())
		return err
	}
	defer server.Shutdown() // 确保在函数返回时关闭服务器
	
	// 添加 panic 捕获层，记录 panic 信息并向上抛出
	defer func() {
		if x := recover(); x != nil {
			var buf bytes.Buffer
			// 导出所有 goroutine 的堆栈信息
			pprof.Lookup("goroutine").WriteTo(&buf, 2)
			mlog.Error("A panic occurred",
				mlog.Any("error", x),
				mlog.String("stack", buf.String()))
			panic(x)
		}
	}()

	// 初始化 REST API 路由
	_, err = api4.Init(server)
	if err != nil {
		mlog.Error(err.Error())
		return err
	}

	// 初始化 WebSocket API
	wsapi.Init(server)
	// 初始化 Web 服务器（静态文件、OAuth、Webhook 等）
	web.New(server)

	// 启动服务器
	err = server.Start()
	if err != nil {
		mlog.Error(err.Error())
		return err
	}

	// 通知系统服务器已就绪
	notifyReady()

	// 清除之前设置的任何信号处理器
	// 这可能是因为某些中间信号处理器需要在 server.Start 完成前清理资源
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	// 等待终止信号，然后优雅地关闭服务
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan // 阻塞，直到收到中断信号

	return nil
}

// notifyReady 如果运行在 systemd 环境中，通知 systemd 服务器已就绪
func notifyReady() {
	// 检查是否存在 systemd 通知 socket
	systemdSocket := os.Getenv("NOTIFY_SOCKET")
	if systemdSocket != "" {
		mlog.Info("Sending systemd READY notification.")

		err := sendSystemdReadyNotification(systemdSocket)
		if err != nil {
			mlog.Error(err.Error())
		}
	}
}

// sendSystemdReadyNotification 向 systemd 发送 READY=1 消息
func sendSystemdReadyNotification(socketPath string) error {
	msg := "READY=1"
	addr := &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	}
	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(msg))
	return err
}
