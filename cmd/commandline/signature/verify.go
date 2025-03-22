package signature

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func Verify(difypkgPath string, publicKeyPath string) {
	// read the plugin
	plugin, err := os.ReadFile(difypkgPath)
	if err != nil {
		log.Error("Failed to read plugin file: %v", err)
		return
	}

	decoderInstance, err := decoder.NewZipPluginDecoder(plugin)
	if err != nil {
		log.Error("Failed to create plugin decoder, plugin path: %s, error: %v", difypkgPath, err)
		return
	}

	if publicKeyPath == "" {
		// verify the plugin with the official (bundled) public key
		err = decoder.VerifyPlugin(decoderInstance)
		if err != nil {
			log.Error("Failed to verify plugin with official public key: %v", err)
			return
		}
	} else {
		// read the public key
		publicKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			log.Error("Failed to read public key file: %v", err)
			return
		}

		publicKey, err := encryption.LoadPublicKey(publicKeyBytes)
		if err != nil {
			log.Error("Failed to load public key: %v", err)
			return
		}

		// verify the plugin
		err = decoder.VerifyPluginWithPublicKey(decoderInstance, publicKey)
		if err != nil {
			log.Error("Failed to verify plugin with provided public key: %v", err)
			return
		}
	}

	log.Info("Plugin verified successfully")
}
