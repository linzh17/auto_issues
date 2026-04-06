# Changelog

## 2026-04-06

### Issue #11: 支持设置工作目录
- 新增 `-workdir` 命令行参数
- 默认使用 `os.Getwd()` 获取当前程序执行目录
- 移除了硬编码的工作目录路径

### Issue #7: 更新 README.md 文档
- 更新 README.md，添加 `-workdir` 参数说明
- 添加 Changelog 章节引用
