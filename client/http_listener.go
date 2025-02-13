package client

import (
	"sync_server/share"

	"github.com/labstack/echo/v4"
)

type HttpServer struct {
	Cfg *share.ClientConfig
}

func NewHttpListener(cfg *share.ClientConfig) *HttpServer {
	return &HttpServer{Cfg: cfg}
}

func (h *HttpServer) Listen() error {
	e := echo.New()
	syncGroup := e.Group("sync-dirs")
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"ClientId":     h.Cfg.ClientId,
			"SyncDirs":     h.Cfg.SyncDirs,
			"SyncInterval": h.Cfg.SyncInterval,
		})
	})
	syncGroup.POST("/", func(c echo.Context) error {
		dirs := h.Cfg.SyncDirs
		newDirs := append(dirs, c.QueryParam("dir"))
		h.Cfg.SyncDirs = newDirs
		share.WriteClientConfig()
		return c.JSON(200, map[string][]string{
			"dirs": newDirs,
		})
	})
	syncGroup.DELETE("/", func(c echo.Context) error {
		dirs := h.Cfg.SyncDirs
		newDirs := remove(dirs, c.QueryParam("dir"))
		h.Cfg.SyncDirs = newDirs
		share.WriteClientConfig()
		return c.JSON(200, map[string][]string{
			"dirs": newDirs,
		})
	})
	return e.Start(":" + h.Cfg.HttpPort)
}
func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
