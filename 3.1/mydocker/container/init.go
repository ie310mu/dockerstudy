package container

import (
	"os"
	"syscall"
	"github.com/Sirupsen/logrus"
)

//容器初始化
func RunContainerInitProcess(command string, args []string) error {
	logrus.Infof("command %s", command)

	//systemd加入linux之后，mount namespace就普成shared by default，所以你必须显式声明你要这个新的mount namesapce独立
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")  //挂载 proc 文件系统，以便后面通过 ps 等系统命令去查看当前进程资源 情况。
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {  //以指定的程序替换当前进程（最开始fork出来的是一个新的mydocker进程，要替换为mydocker run xxx命令指定的程序）
		logrus.Errorf(err.Error())
	}
	return nil
}
