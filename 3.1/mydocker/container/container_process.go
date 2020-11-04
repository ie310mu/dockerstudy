package container

import (
	"syscall"
	"os/exec"
	"os"
)

//以自身fork进程，并调用init命令
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}  //指定容器启动时调用init命令
	cmd := exec.Command("/proc/self/exe", args...)  
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
		syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
    }
	if tty {  //指定了-ti时使用系统输入输出
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
