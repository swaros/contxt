package yacl

// SupportMigrate apply a callback they will be executed if the config is loaded somehow.
// this means for any config that is loaded by the defined rules.
// that is different to the config that is used at the end,
// because yacl can loads a couple of config files and just use the last one.

func (c *ConfigModel) SupportMigrate(fileHandelFn func(path string, cfg interface{})) *ConfigModel {
	c.fileLoadCallback = fileHandelFn
	c.supportMigrate = true
	return c
}
