package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func main() {
	done := make(chan struct{})
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K") // 回到行首并清除该行
				return
			default:
				fmt.Printf("\r%s 加载中...", frames[i%len(frames)])
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	info, err := GetTodayWorkInfo()
	close(done)
	time.Sleep(100 * time.Millisecond) // 等最后一帧结束，避免闪烁

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// 今天是礼拜X
	msg := "今天是" + info.WeekdayName
	// 非周末则加：距离周末还有X天
	if info.DaysToWeekend > 0 {
		msg += fmt.Sprintf("，距离周末还有%d天，", info.DaysToWeekend)
	}
	// 距离最近节假日多少天
	if info.NearestName != "" {
		msg += fmt.Sprintf("距离最近节假日（%s）还有%d天，", info.NearestName, info.DaysToHoliday)
	} else {
		msg += "暂无后续节假日信息。"
	}
	fmt.Println(msg)

	// 按任意键退出，避免命令行窗口直接关闭
	waitKeyToExit()
}

// waitKeyToExit 等待用户按任意键后退出（Windows 用 cmd pause，其它系统用读一行）
func waitKeyToExit() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "pause")
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		return
	}
	fmt.Print("按 Enter 键退出...")
	var b [1]byte
	_, _ = os.Stdin.Read(b[:])
}
