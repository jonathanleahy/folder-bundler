package compression

import (
	"github.com/jonathanleahy/folder-bundler/internal/compression/adapters"
)

// InitializeStrategies registers all available compression strategies
func InitializeStrategies() error {
	// Register none (passthrough) strategy
	if err := DefaultRegistry.Register(adapters.NewNoneCompression()); err != nil {
		return err
	}
	
	// Register dictionary compression
	if err := DefaultRegistry.Register(adapters.NewDictionaryCompression()); err != nil {
		return err
	}
	
	// Register template compression
	if err := DefaultRegistry.Register(adapters.NewTemplateCompression()); err != nil {
		return err
	}
	
	// Register delta compression
	if err := DefaultRegistry.Register(adapters.NewDeltaCompression()); err != nil {
		return err
	}
	
	// Register combined template+delta compression
	if err := DefaultRegistry.Register(adapters.NewTemplateDeltaCompression()); err != nil {
		return err
	}
	
	// Future strategies can be registered here
	// - Run-length encoding
	
	return nil
}