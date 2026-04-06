package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	skillName    = "auto-issue-v2"
	skillBaseDir = "skills"
)

func main() {
	interval := flag.Duration("interval", 30*time.Minute, "执行间隔，如 30s, 30m, 30h")
	prompt := flag.String("prompt", "在 jihulab 上处理 linzh17-group/linzh17-project 的 open issues", "传递给 amp 的 prompt")
	workDir := flag.String("workdir", "", "工作目录，默认为当前程序执行目录")
	maxConcurrency := flag.Int("concurrency", 5, "最大并发任务数")
	autoInstall := flag.Bool("auto-install", false, "自动安装技能到 Amp 技能目录（无需询问）")
	installPath := flag.String("install-path", "", "指定技能安装路径（默认 ~/.config/agents/skills/）")
	flag.Parse()

	// 检查并提示安装技能（仅首次运行时）
	checkAndInstallSkill(*autoInstall, *installPath)

	// 初始化并发控制
	semaphore := make(chan struct{}, *maxConcurrency)
	var wg sync.WaitGroup

	// 首次执行
	runTaskAsync(&wg, semaphore, *prompt, *workDir)

	// 定时执行
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	fmt.Printf("[%s] 定时任务已启动，每 %s 执行一次，最大并发数: %d\n", time.Now().Format("2006-01-02 15:04:05"), *interval, *maxConcurrency)

	for range ticker.C {
		runTaskAsync(&wg, semaphore, *prompt, *workDir)
	}

	// 等待所有任务完成
	wg.Wait()
}

// getSkillPaths 返回可能的技能安装目录路径
func getSkillPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".config", "agents", "skills"),
		filepath.Join(home, ".config", "amp", "skills"),
		".agents/skills",
		".claude/skills",
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".cursor", "skills"),
	}
}

// checkAndInstallSkill 检查技能是否已安装，如未安装则提示用户安装
func checkAndInstallSkill(autoInstall bool, customPath string) {
	// 检查是否已安装
	if isSkillInstalled() {
		return
	}

	// 确定安装路径
	installDir := customPath
	if installDir == "" {
		home, _ := os.UserHomeDir()
		installDir = filepath.Join(home, ".config", "agents", "skills")
	}

	// 自动安装或询问用户
	if autoInstall {
		installSkill(installDir)
	} else {
		fmt.Printf("\n🔍 检测到技能 '%s' 未安装。\n", skillName)
		fmt.Printf("📁 默认安装路径: %s\n\n", installDir)
		fmt.Print("是否安装技能？[Y/n]: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" || input == "y" || input == "yes" {
			installSkill(installDir)
		} else {
			fmt.Println("跳过技能安装。继续执行...\n")
		}
	}
}

// isSkillInstalled 检查技能是否已安装在任意可能的路径
func isSkillInstalled() bool {
	for _, path := range getSkillPaths() {
		skillPath := filepath.Join(path, skillName, "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			return true
		}
	}
	return false
}

// installSkill 将技能安装到指定目录
func installSkill(installDir string) {
	// 获取当前程序所在目录
	exePath, err := os.Executable()
	if err != nil {
		exePath, _ = os.Getwd()
	}

	// 技能源目录（相对于可执行文件或工作目录）
	sourceSkillDir := filepath.Join(filepath.Dir(exePath), skillBaseDir, skillName)

	// 如果源目录不存在，尝试相对于工作目录
	if _, err := os.Stat(sourceSkillDir); os.IsNotExist(err) {
		sourceSkillDir = filepath.Join("skills", skillName)
	}

	// 目标目录
	targetSkillDir := filepath.Join(installDir, skillName)

	// 检查源目录是否存在
	if _, err := os.Stat(sourceSkillDir); os.IsNotExist(err) {
		fmt.Printf("⚠️  技能源目录不存在: %s\n", sourceSkillDir)
		fmt.Println("请确保技能文件存在于项目的 skills/auto-issue-v2/ 目录下。")
		return
	}

	// 创建目标目录
	if err := os.MkdirAll(installDir, 0755); err != nil {
		fmt.Printf("⚠️  创建目录失败: %v\n", err)
		return
	}

	// 复制技能文件
	if err := copyDir(sourceSkillDir, targetSkillDir); err != nil {
		fmt.Printf("⚠️  安装技能失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 技能已成功安装到: %s\n", targetSkillDir)
	fmt.Println("请重新加载 Amp 以使用新安装的技能。\n")
}

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

var execCommand = exec.Command

func runTaskAsync(wg *sync.WaitGroup, semaphore chan struct{}, prompt string, workDir string) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		// 获取信号量
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		executeTask(prompt, workDir)
	}()
}

func executeTask(prompt string, workDir string) {
	fmt.Printf("[%s] 开始执行任务...\n", time.Now().Format("2006-01-02 15:04:05"))

	cmd := execCommand("amp", "-x",
		prompt,
		"--dangerously-allow-all")

	// 设置环境变量
	cmd.Env = append(os.Environ(),
		"AMP_URL=http://localhost:8317",
		"AMP_API_KEY=your-api-key-1",
	)

	// 设置工作目录：如果未指定，则使用当前程序执行目录
	if workDir != "" {
		cmd.Dir = workDir
	} else {
		if cwd, err := os.Getwd(); err == nil {
			cmd.Dir = cwd
		}
	}

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[%s] 执行失败: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
	}

	fmt.Printf("[%s] 输出:\n%s\n", time.Now().Format("2006-01-02 15:04:05"), string(output))
}
