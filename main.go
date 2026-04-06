package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	interval := flag.Duration("interval", 30*time.Minute, "执行间隔，如 30s, 30m, 30h")
	flag.Parse()

	// 首次执行
	runTask()

	// 定时执行
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	fmt.Printf("[%s] 定时任务已启动，每 %s 执行一次\n", time.Now().Format("2006-01-02 15:04:05"), *interval)

	for range ticker.C {
		runTask()
	}
}

var execCommand = exec.Command

func runTask() {
	fmt.Printf("[%s] 开始执行任务...\n", time.Now().Format("2006-01-02 15:04:05"))

	cmd := execCommand("amp", "-x",
		"在 jihulab 上处理 linzh17-group/linzh17-project 的 open issues",
		"--dangerously-allow-all")

	// 设置环境变量
	cmd.Env = append(os.Environ(),
		"AMP_URL=http://localhost:8317",
		"AMP_API_KEY=your-api-key-1",
	)

	// 设置工作目录（如果需要）
	cmd.Dir = "/Users/lzh17/Projects/gitlab_auto_test/corn_amp"

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[%s] 执行失败: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
	}

	fmt.Printf("[%s] 输出:\n%s\n", time.Now().Format("2006-01-02 15:04:05"), string(output))
}


