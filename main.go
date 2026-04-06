package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

func main() {
	interval := flag.Duration("interval", 30*time.Minute, "执行间隔，如 30s, 30m, 30h")
	prompt := flag.String("prompt", "在 jihulab 上处理 linzh17-group/linzh17-project 的 open issues", "传递给 amp 的 prompt")
	workDir := flag.String("workdir", "", "工作目录，默认为当前程序执行目录")
	maxConcurrency := flag.Int("concurrency", 5, "最大并发任务数")
	flag.Parse()

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
