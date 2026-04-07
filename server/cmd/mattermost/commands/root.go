// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// 根命令定义，使用 Cobra 框架构建 CLI 接口
package commands

import (
	"os"

	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/spf13/cobra"
)

// Command 为 cobra.Command 的类型别名，方便使用
type Command = cobra.Command

// Run 执行根命令并传入参数
func Run(args []string) error {
	RootCmd.SetArgs(args)
	return RootCmd.Execute()
}

// RootCmd 定义 Mattermost 的根命令
var RootCmd = &cobra.Command{
	Use:   "mattermost",
	Short: "Open source, self-hosted Slack-alternative",
	Long:  `Mattermost offers workplace messaging across web, PC and phones with archiving, search and integration with your existing systems. Documentation available at https://docs.mattermost.com`,
	// 在每个子命令运行之前执行
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		checkForRootUser()
	},
}

func init() {
	// 添加全局标志 --config / -c，用于指定配置文件路径
	RootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file to use.")
}

// checkForRootUser 检查当前用户是否为 root，如果是则记录警告日志
// 以 root 运行 Mattermost 不被推荐，因为存在安全风险
func checkForRootUser() {
	if os.Geteuid() == 0 {
		mlog.Warn("Running Mattermost as root is not recommended. Please use a non-root user.")
	}
}
