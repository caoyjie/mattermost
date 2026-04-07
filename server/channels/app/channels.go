// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// Channels 是频道相关业务逻辑的核心容器
// 包含所有与频道相关的状态和服务
package app

import (
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/channels/app/imaging"
	"github.com/mattermost/mattermost/server/v8/config"
	"github.com/mattermost/mattermost/server/v8/einterfaces"
	"github.com/mattermost/mattermost/server/v8/platform/services/imageproxy"
	"github.com/mattermost/mattermost/server/v8/platform/shared/filestore"
)

// configService 配置服务接口，用于管理和更新配置
type configService interface {
	Config() *model.Config
	AddConfigListener(listener func(*model.Config, *model.Config)) string
	RemoveConfigListener(id string)
	UpdateConfig(f func(*model.Config))
	SaveConfig(newCfg *model.Config, sendConfigChangeClusterMessage bool) (*model.Config, *model.Config, *model.AppError)
}

// Channels 包含所有频道相关的状态
type Channels struct {
	srv             *Server                     // 引用 Server 实例
	cfgSvc          configService               // 配置服务
	filestore       filestore.FileBackend       // 文件存储后端
	exportFilestore filestore.FileBackend       // 导出文件存储后端

	postActionCookieSecret []byte // 帖子操作 Cookie 密钥

	pluginCommandsLock            sync.RWMutex   // 插件命令锁
	pluginCommands                []*PluginCommand // 插件命令列表
	pluginsLock                   sync.RWMutex   // 插件锁
	pluginsEnvironment            *plugin.Environment // 插件环境
	pluginConfigListenerID        string          // 插件配置监听器 ID
	pluginClusterLeaderListenerID string          // 插件集群领导者监听器 ID

	imageProxy *imageproxy.ImageProxy // 图片代理服务

	// 以下缓存计数用于通知条件验证
	cachedPostCount   int64 // 缓存的帖子数量
	cachedUserCount   int64 // 缓存的用户数量
	cachedDBMSVersion string // 缓存的数据库版本
	// 之前获取的通知
	cachedNotices model.ProductNotices

	// 企业版接口
	AccountMigration einterfaces.AccountMigrationInterface
	Compliance       einterfaces.ComplianceInterface
	DataRetention    einterfaces.DataRetentionInterface
	MessageExport    einterfaces.MessageExportInterface
	Saml             einterfaces.SamlInterface
	Notification     einterfaces.NotificationInterface
	Ldap             einterfaces.LdapInterface
	AccessControl    einterfaces.AccessControlServiceInterface
	Intune           einterfaces.IntuneInterface

	// 以下用于防止并发上传请求导致的不一致和数据损坏
	uploadLockMapMut sync.Mutex
	uploadLockMap    map[string]bool

	imgDecoder *imaging.Decoder // 图片解码器
	imgEncoder *imaging.Encoder // 图片编码器

	dndTaskMut sync.Mutex
	dndTask    *model.ScheduledTask // "请勿打扰" 定时任务

	postReminderMut  sync.Mutex
	postReminderTask *model.ScheduledTask // 帖子提醒任务

	interruptQuitChan     chan struct{}
	scheduledPostMut      sync.Mutex
	scheduledPostTask     *model.ScheduledTask // 定时发帖任务
	emailLoginAttemptsMut sync.Mutex
	ldapLoginAttemptsMut  sync.Mutex
}

// NewChannels 创建并初始化 Channels 实例
func NewChannels(s *Server) (*Channels, error) {
	ch := &Channels{
		srv:               s,
		imageProxy:        imageproxy.MakeImageProxy(s.platform, s.httpService, s.Log()),
		uploadLockMap:     map[string]bool{},
		filestore:         s.FileBackend(),
		exportFilestore:   s.ExportFileBackend(),
		cfgSvc:            s.Platform(),
		interruptQuitChan: make(chan struct{}),
	}

	// We are passing a partially filled Channels struct so that the enterprise
	// methods can have access to app methods.
	// Otherwise, passing server would mean it has to call s.Channels(),
	// which would be nil at this point.
	if complianceInterface != nil {
		ch.Compliance = complianceInterface(New(ServerConnector(ch)))
	}
	if messageExportInterface != nil {
		ch.MessageExport = messageExportInterface(New(ServerConnector(ch)))
	}
	if dataRetentionInterface != nil {
		ch.DataRetention = dataRetentionInterface(New(ServerConnector(ch)))
	}
	if accountMigrationInterface != nil {
		ch.AccountMigration = accountMigrationInterface(New(ServerConnector(ch)))
	}
	if ldapInterface != nil {
		ch.Ldap = ldapInterface(New(ServerConnector(ch)))
	}
	if notificationInterface != nil {
		ch.Notification = notificationInterface(New(ServerConnector(ch)))
	}
	if samlInterface != nil {
		ch.Saml = samlInterface(New(ServerConnector(ch)))
		if err := ch.Saml.ConfigureSP(request.EmptyContext(s.Log())); err != nil {
			s.Log().Error("An error occurred while configuring SAML Service Provider", mlog.Err(err))
		}

		ch.AddConfigListener(func(_, _ *model.Config) {
			if err := ch.Saml.ConfigureSP(request.EmptyContext(s.Log())); err != nil {
				s.Log().Error("An error occurred while configuring SAML Service Provider", mlog.Err(err))
			}
		})
	}

	if intuneInterface != nil {
		ch.Intune = intuneInterface(New(ServerConnector(ch)))
	}

	if pushProxyInterface != nil {
		app := New(ServerConnector(ch))
		s.PushProxy = pushProxyInterface(app)

		// Add config listener to regenerate token when push proxy URL changes
		app.AddConfigListener(func(oldCfg, newCfg *model.Config) {
			// Only cluster leader should regenerate to avoid duplicate requests
			if !app.IsLeader() {
				return
			}

			oldURL := model.SafeDereference(oldCfg.EmailSettings.PushNotificationServer)
			newURL := model.SafeDereference(newCfg.EmailSettings.PushNotificationServer)

			// If push proxy URL changed
			if oldURL != newURL {
				if newURL != "" {
					// URL changed to a new value, regenerate token
					s.Log().Info("Push notification server URL changed, regenerating auth token",
						mlog.String("old_url", oldURL),
						mlog.String("new_url", newURL))

					if err := s.PushProxy.GenerateAuthToken(); err != nil {
						s.Log().Error("Failed to regenerate auth token after config change", mlog.Err(err))
					}
				} else if oldURL != "" {
					// URL was cleared, delete the old token
					s.Log().Info("Push notification server URL cleared, removing auth token")
					if err := s.PushProxy.DeleteAuthToken(); err != nil {
						s.Log().Error("Failed to delete auth token after URL cleared", mlog.Err(err))
					}
				}
			}
		})
	}

	if accessControlServiceInterface != nil {
		app := New(ServerConnector(ch))
		ch.AccessControl = accessControlServiceInterface(app)

		appErr := ch.AccessControl.Init(request.EmptyContext(s.Log()))
		if appErr != nil && appErr.StatusCode != http.StatusNotImplemented {
			s.Log().Error("An error occurred while initializing Access Control", mlog.Err(appErr))
		}

		app.AddLicenseListener(func(newCfg, old *model.License) {
			if ch.AccessControl != nil {
				if appErr := ch.AccessControl.Init(request.EmptyContext(s.Log())); appErr != nil && appErr.StatusCode != http.StatusNotImplemented {
					s.Log().Error("An error occurred while initializing Access Control", mlog.Err(appErr))
				}
			}
		})
	}

	var imgErr error
	decoderConcurrency := int(*ch.cfgSvc.Config().FileSettings.MaxImageDecoderConcurrency)
	if decoderConcurrency == -1 {
		decoderConcurrency = runtime.NumCPU()
	}
	ch.imgDecoder, imgErr = imaging.NewDecoder(imaging.DecoderOptions{
		ConcurrencyLevel: decoderConcurrency,
	})
	if imgErr != nil {
		return nil, errors.Wrap(imgErr, "failed to create image decoder")
	}
	ch.imgEncoder, imgErr = imaging.NewEncoder(imaging.EncoderOptions{
		ConcurrencyLevel: runtime.NumCPU(),
	})
	if imgErr != nil {
		return nil, errors.Wrap(imgErr, "failed to create image encoder")
	}

	// Setup routes.
	pluginsRoute := ch.srv.Router.PathPrefix("/plugins/{plugin_id:[A-Za-z0-9\\_\\-\\.]+}").Subrouter()
	pluginsRoute.HandleFunc("", ch.ServePluginRequest)
	pluginsRoute.HandleFunc("/public/{public_file:.*}", ch.ServePluginPublicRequest)
	pluginsRoute.HandleFunc("/{anything:.*}", ch.ServePluginRequest)

	return ch, nil
}

func (ch *Channels) Start() error {
	// Start plugins
	ctx := request.EmptyContext(ch.srv.Log())
	ch.initPlugins(ctx, *ch.cfgSvc.Config().PluginSettings.Directory, *ch.cfgSvc.Config().PluginSettings.ClientDirectory)

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-interruptChan:
			if err := ch.Stop(); err != nil {
				ch.srv.Log().Warn("Error stopping channels", mlog.Err(err))
			}
			os.Exit(1)
		case <-ch.interruptQuitChan:
			return
		}
	}()

	ch.AddConfigListener(func(prevCfg, cfg *model.Config) {
		// We compute the difference between configs
		// to ensure we don't re-init plugins unnecessarily.
		diffs, err := config.Diff(prevCfg, cfg)
		if err != nil {
			ch.srv.Log().Warn("Error in comparing configs", mlog.Err(err))
			return
		}

		hasDiff := false
		// TODO: This could be a method on ConfigDiffs itself
		for _, diff := range diffs {
			if strings.HasPrefix(diff.Path, "PluginSettings.") {
				hasDiff = true
				break
			}
		}

		// Do only if some plugin related settings has changed.
		if hasDiff {
			if *cfg.PluginSettings.Enable {
				ch.initPlugins(ctx, *cfg.PluginSettings.Directory, *ch.cfgSvc.Config().PluginSettings.ClientDirectory)
			} else {
				ch.ShutDownPlugins()
			}
		}
	})

	// TODO: This should be moved to the platform service.
	if err := ch.srv.platform.EnsureAsymmetricSigningKey(); err != nil {
		return errors.Wrapf(err, "unable to ensure asymmetric signing key")
	}

	if err := ch.ensurePostActionCookieSecret(); err != nil {
		return errors.Wrapf(err, "unable to ensure PostAction cookie secret")
	}

	return nil
}

func (ch *Channels) Stop() error {
	ch.ShutDownPlugins()

	ch.dndTaskMut.Lock()
	if ch.dndTask != nil {
		ch.dndTask.Cancel()
	}
	ch.dndTaskMut.Unlock()

	close(ch.interruptQuitChan)

	return nil
}

func (ch *Channels) AddConfigListener(listener func(*model.Config, *model.Config)) string {
	return ch.cfgSvc.AddConfigListener(listener)
}

func (ch *Channels) RemoveConfigListener(id string) {
	ch.cfgSvc.RemoveConfigListener(id)
}

func (ch *Channels) RunMultiHook(hookRunnerFunc func(hooks plugin.Hooks, manifest *model.Manifest) bool, hookId int) {
	if env := ch.GetPluginsEnvironment(); env != nil {
		env.RunMultiPluginHook(hookRunnerFunc, hookId)
	}
}

func (ch *Channels) HooksForPlugin(id string) (plugin.Hooks, error) {
	env := ch.GetPluginsEnvironment()
	if env == nil {
		return nil, errors.New("plugins are not initialized")
	}

	hooks, err := env.HooksForPlugin(id)
	if err != nil {
		return nil, err
	}

	return hooks, nil
}
