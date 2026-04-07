// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// Mattermost 服务器主入口点
// 负责初始化并启动整个 Mattermost 服务
package main

import (
	"os"

	"github.com/mattermost/mattermost/server/v8/cmd/mattermost/commands"
	// 导入并注册应用层的斜杠命令（如 /me, /away 等）
	_ "github.com/mattermost/mattermost/server/v8/channels/app/slashcommands"
	// 导入 OAuth 提供者（GitLab）
	_ "github.com/mattermost/mattermost/server/v8/channels/app/oauthproviders/gitlab"

	// 导入企业版功能模块（LDAP、SAML、合规性等）
	_ "github.com/mattermost/mattermost/server/v8/enterprise"
)

func main() {
	// 执行 CLI 命令，如果发生错误则退出码为 1
	if err := commands.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
