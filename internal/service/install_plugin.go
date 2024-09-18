package service

import (
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InstallPluginFromPkg(c *gin.Context, tenant_id string, dify_pkg_file multipart.File) {
	manager := plugin_manager.Manager()

	plugin_file, err := io.ReadAll(dify_pkg_file)
	if err != nil {
		c.JSON(200, entities.NewErrorResponse(-500, err.Error()))
		return
	}

	decoder, err := decoder.NewZipPluginDecoder(plugin_file)
	if err != nil {
		c.JSON(200, entities.NewErrorResponse(-500, err.Error()))
		return
	}

	baseSSEService(
		func() (*stream.Stream[plugin_manager.PluginInstallResponse], error) {
			return manager.Install(tenant_id, decoder)
		},
		c,
		3600,
	)
}

func InstallPluginFromIdentifier(
	c *gin.Context,
	tenant_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	// check if identifier exists
	plugin, err := db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return entities.NewErrorResponse(-404, "Plugin not found")
	}
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	declaration, err := plugin.GetDeclaration()
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	// install to this workspace
	if _, _, err := curd.CreatePlugin(tenant_id, plugin_unique_identifier, plugin.InstallType, declaration); err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(plugin)
}