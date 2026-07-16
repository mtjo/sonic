package theme

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"go.uber.org/fx"

	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type gitThemeFetcherImpl struct {
	fx.Out
	PropertyScanner PropertyScanner
}

func (g gitThemeFetcherImpl) FetchTheme(ctx context.Context, file interface{}) (*dto.ThemeProperty, error) {
	gitURL := file.(string)
	// 取 URL 最后一段作为临时目录名，去掉尾部斜杠避免得到空名（空名会令
	// filepath.Join(TempDir, "") == TempDir，从而误删整个系统临时目录）
	themeDirName := strings.TrimRight(gitURL, "/")
	if idx := strings.LastIndex(themeDirName, "/"); idx >= 0 {
		themeDirName = themeDirName[idx+1:]
	}
	themeDirName = strings.TrimSuffix(themeDirName, ".git")
	themeDirName = strings.Trim(themeDirName, ".")
	if themeDirName == "" {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("invalid theme url")
	}
	tempDir := os.TempDir()
	tmpThemeDir := filepath.Join(tempDir, themeDirName)
	if util.FileIsExisted(tmpThemeDir) {
		err := os.RemoveAll(tmpThemeDir)
		if err != nil {
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("delete tmp theme directory err")
		}
	}
	_, err := git.PlainClone(filepath.Join(tempDir, themeDirName), false, &git.CloneOptions{
		URL: gitURL,
	})
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(err.Error())
	}
	// 移除克隆下来的 .git 目录，避免被拷入主题目录
	_ = os.RemoveAll(filepath.Join(tempDir, themeDirName, ".git"))
	themeProperty, err := g.PropertyScanner.ReadThemeProperty(ctx, filepath.Join(tempDir, themeDirName))
	if err != nil {
		return nil, err
	}
	return themeProperty, nil
}

func NewGitThemeFetcher(propertyScanner PropertyScanner) ThemeFetcher {
	return &gitThemeFetcherImpl{
		PropertyScanner: propertyScanner,
	}
}
