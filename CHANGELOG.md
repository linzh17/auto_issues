# Changelog

## 2026-04-06

### Issue #11: 支持设置工作目录
- 新增 `-workdir` 命令行参数
- 默认使用 `os.Getwd()` 获取当前程序执行目录
- 移除了硬编码的工作目录路径

### Issue #7: 更新 README.md 文档
- 更新 README.md，添加 `-workdir` 参数说明
- 添加 Changelog 章节引用
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2026-04-06

### Added
- SKILL.md 工作流程新增 changelog 更新要求：每次代码改动必须更新对应项目的 changelog

## 2026-04-06 - Issue #14

### Feature: 运行时自动加载本技能
- 新增 `-auto-install` 参数，支持自动安装技能（无需询问）
- 新增 `-install-path` 参数，允许指定技能安装路径
- 首次运行时自动检测技能是否已安装
- 若未安装，提示用户确认安装路径（默认 ~/.config/agents/skills/）
- 自动创建缺失的目录结构
- 支持多个 Amp 技能目录：~/.config/agents/skills/, ~/.config/amp/skills/, .agents/skills/ 等

## 2026-04-06 - Issue #16

### Feature: 定时任务支持并发执行
- 新增 `-concurrency` 命令行参数控制最大并发数
- 使用带缓冲的 channel 作为信号量实现并发控制
- 定时触发时启动异步协程执行任务，主协程直接返回继续调度
- 使用 sync.WaitGroup 确保程序退出前所有任务完成
