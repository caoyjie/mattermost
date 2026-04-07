// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// App 是纯函数式业务逻辑层
// 它不包含任何字段（除了 Server），每次请求时都会创建一个新的 App 实例
// 通过其方法为 Server 提供业务逻辑支持
package app

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/httpservice"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/timezones"
	"github.com/mattermost/mattermost/server/v8/einterfaces"
	"github.com/mattermost/mattermost/server/v8/platform/services/imageproxy"
	"github.com/mattermost/mattermost/server/v8/platform/services/searchengine"
	"github.com/mattermost/mattermost/server/v8/platform/shared/templates"
)

// App 是纯函数式组件，不包含任何字段（除了 Server）
// 它是一个请求级别的结构体，每次请求都会创建新实例
// 唯一目的是通过其方法为 Server 提供业务逻辑
type App struct {
	ch *Channels // 引用 Channels 实例
}

// New 创建一个新的 App 实例，并应用所有传入的选项
func New(options ...AppOption) *App {
	app := &App{}

	for _, option := range options {
		option(app)
	}

	return app
}

// ServerId 返回服务器 ID
func (a *App) ServerId() string {
	return a.Srv().ServerId()
}

// TemplatesContainer 返回 HTML 模板容器
func (s *Server) TemplatesContainer() *templates.Container {
	return s.htmlTemplates
}

// getFirstServerRunTimestamp 获取服务器首次运行时间戳
func (s *Server) getFirstServerRunTimestamp() (int64, *model.AppError) {
	systemData, err := s.Store().System().GetByName(model.SystemFirstServerRunTimestampKey)
	if err != nil {
		return 0, model.NewAppError("getFirstServerRunTimestamp", "app.system.get_by_name.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	value, err := strconv.ParseInt(systemData.Value, 10, 64)
	if err != nil {
		return 0, model.NewAppError("getFirstServerRunTimestamp", "app.system_install_date.parse_int.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return value, nil
}

// Channels 返回 Channels 实例
func (a *App) Channels() *Channels {
	return a.ch
}

// Srv 返回 Server 实例
func (a *App) Srv() *Server {
	return a.ch.srv
}

// Log 返回日志记录器
func (a *App) Log() *mlog.Logger {
	return a.ch.srv.Log()
}

// 以下方法提供各种企业接口的访问
// 这些接口在开源版本和企业版中可能有不同的实现

// AccountMigration 返回账户迁移接口
func (a *App) AccountMigration() einterfaces.AccountMigrationInterface {
	return a.ch.AccountMigration
}

// Cluster 返回集群接口
func (a *App) Cluster() einterfaces.ClusterInterface {
	return a.ch.srv.platform.Cluster()
}

// Compliance 返回合规性接口
func (a *App) Compliance() einterfaces.ComplianceInterface {
	return a.ch.Compliance
}

// DataRetention 返回数据保留接口
func (a *App) DataRetention() einterfaces.DataRetentionInterface {
	return a.ch.DataRetention
}

// SearchEngine 返回搜索引擎接口
func (a *App) SearchEngine() *searchengine.Broker {
	return a.ch.srv.platform.SearchEngine
}

// Ldap 返回 LDAP 接口
func (a *App) Ldap() einterfaces.LdapInterface {
	return a.ch.Ldap
}

// LdapDiagnostic 返回 LDAP 诊断接口
func (a *App) LdapDiagnostic() einterfaces.LdapDiagnosticInterface {
	return a.ch.srv.platform.LdapDiagnostic()
}

// MessageExport 返回消息导出接口
func (a *App) MessageExport() einterfaces.MessageExportInterface {
	return a.ch.MessageExport
}

// Metrics 返回指标接口
func (a *App) Metrics() einterfaces.MetricsInterface {
	return a.ch.srv.GetMetrics()
}

// Notification 返回通知接口
func (a *App) Notification() einterfaces.NotificationInterface {
	return a.ch.Notification
}

// AutoTranslation 返回自动翻译接口
func (a *App) AutoTranslation() einterfaces.AutoTranslationInterface {
	return a.Srv().AutoTranslation
}
func (a *App) Saml() einterfaces.SamlInterface {
	return a.ch.Saml
}
func (a *App) Intune() einterfaces.IntuneInterface {
	return a.ch.Intune
}
func (a *App) Cloud() einterfaces.CloudInterface {
	return a.ch.srv.Cloud
}
func (a *App) IPFiltering() einterfaces.IPFilteringInterface {
	return a.ch.srv.IPFiltering
}
func (a *App) OutgoingOAuthConnections() einterfaces.OutgoingOAuthConnectionInterface {
	return a.ch.srv.OutgoingOAuthConnection
}
func (a *App) HTTPService() httpservice.HTTPService {
	return a.ch.srv.httpService
}
func (a *App) ImageProxy() *imageproxy.ImageProxy {
	return a.ch.imageProxy
}
func (a *App) Timezones() *timezones.Timezones {
	return a.ch.srv.timezones
}
func (a *App) License() *model.License {
	return a.Srv().License()
}

func (a *App) DBHealthCheckWrite() error {
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)

	return a.Srv().Store().System().SaveOrUpdate(&model.System{
		Name:  a.dbHealthCheckKey(),
		Value: currentTime,
	})
}

func (a *App) DBHealthCheckDelete() error {
	_, err := a.Srv().Store().System().PermanentDeleteByName(a.dbHealthCheckKey())
	return err
}

func (a *App) dbHealthCheckKey() string {
	return fmt.Sprintf("health_check_%s", a.GetClusterId())
}

func (a *App) CheckIntegrity() <-chan model.IntegrityCheckResult {
	return a.Srv().Store().CheckIntegrity()
}

func (a *App) SetChannels(ch *Channels) {
	a.ch = ch
}

func (a *App) SetServer(srv *Server) {
	a.ch.srv = srv
}

func (a *App) PropertyAccessService() *PropertyAccessService {
	return a.Srv().propertyAccessService
}

func (a *App) UpdateExpiredDNDStatuses() ([]*model.Status, error) {
	return a.Srv().Store().Status().UpdateExpiredDNDStatuses()
}
