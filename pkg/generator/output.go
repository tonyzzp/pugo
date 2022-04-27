package generator

import (
	"os"
	"path/filepath"
	"pugo/pkg/models"
	"pugo/pkg/theme"
	"pugo/pkg/utils"
	"pugo/pkg/zlog"
)

// Output outputs contents to destination directory.
func Output(s *models.SiteData, ctx *Context, outputDir string) error {
	if err := updateThemeCopyDirs(s.Render, ctx); err != nil {
		zlog.Warn("theme: failed to update copy dirs", "err", err)
		return err
	}
	if err := outputFiles(s, ctx); err != nil {
		return err
	}
	if err := copyAssets(outputDir, ctx); err != nil {
		return err
	}
	return nil
}

func updateThemeCopyDirs(r *theme.Render, ctx *Context) error {
	staticDirs := r.GetStaticDirs()
	themeDir := r.GetDir()
	for _, dir := range staticDirs {
		ctx.appendCopyDir(filepath.Join(themeDir, dir), dir)
	}
	return nil
}

func outputFiles(s *models.SiteData, ctx *Context) error {
	var err error

	outputs := ctx.GetOutputs()
	for fpath, buf := range outputs {
		data := buf.Bytes()
		dataLen := len(data)
		if s.BuildConfig.EnableMinifyHTML {
			data, err = ctx.MinifyHTML(data)
			if err != nil {
				zlog.Warnf("output: failed to minify: %s, %s", fpath, err)
				data = buf.Bytes()
			} else {
				zlog.Debugf("minified ok: %s, %d -> %d", fpath, dataLen, len(data))
			}
		}
		if err = utils.WriteFile(fpath, data); err != nil {
			zlog.Warnf("output: failed to write file: %s, %s", fpath, err)
			continue
		}
		ctx.incrOutputCounter(1)
	}
	return nil
}

func copyAssets(outputDir string, ctx *Context) error {
	for _, dirData := range ctx.copingDirs {
		err := filepath.Walk(dirData.SrcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			relPath, err := filepath.Rel(dirData.SrcDir, path)
			if err != nil {
				return nil
			}
			dstPath := filepath.Join(dirData.DestDir, relPath)
			dstPath = filepath.Join(outputDir, dstPath)
			if err := utils.CopyFile(path, dstPath); err != nil {
				zlog.Warnf("output: failed to copy file: %s, %s", dstPath, err)
				return err
			}
			zlog.Infof("assets copied: %s", dstPath)
			ctx.incrOutputCounter(1)
			return nil
		})
		if err != nil {
			zlog.Warnf("output: failed to copy assets: %s, %s", dirData.SrcDir, err)
			return err
		}
	}
	return nil
}
