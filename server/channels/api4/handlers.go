// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// handlers.go 定义 API 处理器的认证中间件
// 提供不同类型端点的处理器：无需登录、需要会话、需要云 API 密钥等
package api4

import (
	"net/http"

	"github.com/klauspost/compress/gzhttp"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/v8/channels/app"
	"github.com/mattermost/mattermost/server/v8/channels/web"
)

// Context 是 web.Context 的别名，表示请求上下文
type Context = web.Context

// handlerFunc 定义处理器函数类型
type handlerFunc func(*Context, http.ResponseWriter, *http.Request)

// APIHandlerOption 定义 API 处理器选项类型
type APIHandlerOption string

const (
	handlerParamFileAPI = APIHandlerOption("fileAPI") // 文件 API 选项
)

// APIHandler 为不需要用户登录的 API 端点提供处理器
func (api *API) APIHandler(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:            api.srv,
		HandleFunc:     h,
		HandlerName:    web.GetHandlerName(h),
		RequireSession: false, // 不需要会话
		TrustRequester: false,
		RequireMfa:     false, // 不需要 MFA
		IsStatic:       false,
		IsLocal:        false,
	}
	setHandlerOpts(handler, opts...)

	// 如果启用了 gzip 压缩，则包装处理器
	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

// APISessionRequired 为需要用户登录的 API 端点提供处理器
func (api *API) APISessionRequired(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:            api.srv,
		HandleFunc:     h,
		HandlerName:    web.GetHandlerName(h),
		RequireSession: true, // 需要会话
		TrustRequester: false,
		RequireMfa:     true, // 需要 MFA
		IsStatic:       false,
		IsLocal:        false,
	}
	setHandlerOpts(handler, opts...)

	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

// CloudAPIKeyRequired 为需要从 CWS（Cloud Web Service）访问云安装的 Webhook 端点提供处理器
func (api *API) CloudAPIKeyRequired(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:             api.srv,
		HandleFunc:      h,
		HandlerName:     web.GetHandlerName(h),
		RequireSession:  false,
		RequireCloudKey: true, // 需要云 API 密钥
		TrustRequester:  false,
		RequireMfa:      false,
		IsStatic:        false,
		IsLocal:         false,
	}
	setHandlerOpts(handler, opts...)

	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

// RemoteClusterTokenRequired 为远程集群请求 /remotecluster 端点提供处理器
func (api *API) RemoteClusterTokenRequired(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:                       api.srv,
		HandleFunc:                h,
		HandlerName:               web.GetHandlerName(h),
		RequireSession:            false,
		RequireCloudKey:           false,
		RequireRemoteClusterToken: true, // 需要远程集群令牌
		TrustRequester:            false,
		RequireMfa:                false,
		IsStatic:                  false,
		IsLocal:                   false,
	}
	setHandlerOpts(handler, opts...)

	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

// APISessionRequiredMfa provides a handler for API endpoints which require a logged-in user session  but when accessed,
// if MFA is enabled, the MFA process is not yet complete, and therefore the requirement to have completed the MFA
// authentication must be waived.
func (api *API) APISessionRequiredMfa(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:            api.srv,
		HandleFunc:     h,
		HandlerName:    web.GetHandlerName(h),
		RequireSession: true,
		TrustRequester: false,
		RequireMfa:     false,
		IsStatic:       false,
		IsLocal:        false,
	}
	setHandlerOpts(handler, opts...)

	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

// APIHandlerTrustRequester provides a handler for API endpoints which do not require the user to be logged in and are
// allowed to be requested directly rather than via javascript/XMLHttpRequest, such as site branding images or the
// websocket.
func (api *API) APIHandlerTrustRequester(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:            api.srv,
		HandleFunc:     h,
		HandlerName:    web.GetHandlerName(h),
		RequireSession: false,
		TrustRequester: true,
		RequireMfa:     false,
		IsStatic:       false,
		IsLocal:        false,
	}
	setHandlerOpts(handler, opts...)

	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

// APISessionRequiredTrustRequester provides a handler for API endpoints which do require the user to be logged in and
// are allowed to be requested directly rather than via javascript/XMLHttpRequest, such as emoji or file uploads.
func (api *API) APISessionRequiredTrustRequester(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:            api.srv,
		HandleFunc:     h,
		HandlerName:    web.GetHandlerName(h),
		RequireSession: true,
		TrustRequester: true,
		RequireMfa:     true,
		IsStatic:       false,
		IsLocal:        false,
	}
	setHandlerOpts(handler, opts...)

	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

// DisableWhenBusy provides a handler for API endpoints which should be disabled when the server is under load,
// responding with HTTP 503 (Service Unavailable).
func (api *API) APISessionRequiredDisableWhenBusy(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:             api.srv,
		HandleFunc:      h,
		HandlerName:     web.GetHandlerName(h),
		RequireSession:  true,
		TrustRequester:  false,
		RequireMfa:      true,
		IsStatic:        false,
		IsLocal:         false,
		DisableWhenBusy: true,
	}
	setHandlerOpts(handler, opts...)

	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

// APILocal provides a handler for API endpoints to be used in local
// mode, this is, through a UNIX socket and without an authenticated
// session, but with one that has no user set and no permission
// restrictions
func (api *API) APILocal(h handlerFunc, opts ...APIHandlerOption) http.Handler {
	handler := &web.Handler{
		Srv:            api.srv,
		HandleFunc:     h,
		HandlerName:    web.GetHandlerName(h),
		RequireSession: false,
		TrustRequester: false,
		RequireMfa:     false,
		IsStatic:       false,
		IsLocal:        true,
	}
	setHandlerOpts(handler, opts...)

	if *api.srv.Config().ServiceSettings.WebserverMode == "gzip" {
		return gzhttp.GzipHandler(handler)
	}
	return handler
}

func (api *API) RateLimitedHandler(apiHandler http.Handler, settings model.RateLimitSettings) http.Handler {
	if !*api.srv.Config().RateLimitSettings.Enable {
		return apiHandler
	}

	settings.SetDefaults()

	rateLimiter, err := app.NewRateLimiter(&settings, []string{})
	if err != nil {
		api.srv.Log().Error("getRateLimitedHandler", mlog.Err(err))
		return nil
	}
	return rateLimiter.RateLimitHandler(apiHandler)
}

func requireLicense(c *Context) *model.AppError {
	if c.App.Channels().License() == nil {
		err := model.NewAppError("", "api.license_error", nil, "", http.StatusNotImplemented)
		return err
	}
	return nil
}

func setHandlerOpts(handler *web.Handler, opts ...APIHandlerOption) {
	if len(opts) == 0 {
		return
	}

	for _, option := range opts {
		switch option {
		case handlerParamFileAPI:
			handler.FileAPI = true
		}
	}
}
