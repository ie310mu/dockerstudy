package main

import (
	"dockerstudy/3.1/mydocker/container"
	log "github.com/Sirupsen/logrus"
	"os"
)


func Run(tty bool, command string) {
	parent := container.NewParentProcess(tty, command)  //fork当前进程，fork时会指定init参数
	if err := parent.Start(); err != nil {  //启动，会执行init
		log.Error(err)
	}
	parent.Wait()  //等待
	os.Exit(-1)
}


// package main

// import (
// 	"dockerstudy/3.1/mydocker/subsystems"
// 	log "github.com/Sirupsen/logrus"
// 	"syscall"
// 	"os/exec"
// 	"os"
// )


// func NewParentProcess(tty bool, command string) *exec.Cmd {
//     cmd.SysProcAttr = &syscall.SysProcAttr{
//         Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
// 		syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
//     }
// 	if tty {  //指定了-ti时使用系统输入输出
// 		cmd.Stdin = os.Stdin
// 		cmd.Stdout = os.Stdout
// 		cmd.Stderr = os.Stderr
// 	}
// 	return cmd
// }


// func Run(tty bool, command string, res *subsystems.ResourceConfig) {
// 	parent := NewParentProcess(tty, command) 
// 	if parent == nil{
// 		log.Errorf("New parent process error")
// 		return
// 	}
// 	if err := parent.Start(); err != nil{
// 		log.Error(err)
// 	}

// 	//use mydocker-cgroup as cgroup name 
// 	//创建 cgroup manager ，并通过调用 set apply 设置资源限制并使限制在容器上生效
// 	cgroupManager := subsystems.NewCgroupManager("mydocker-cgroup");
// 	defer cgroupManager.Destory()
// 	//／设置资源限制
// 	cgroupManager.Set(res) 
// 	//／将容器进程加入到各个 subsystem 挂载对应的 cgroup中
// 	cgroupManager.Apply(parent.Process.Pid)
// 	//／对容器设置完限制之后 初始 容器
// 	args := []string{"init", command}  //指定容器启动时调用init命令
// 	cmd := exec.Command("/proc/self/exe", args...)  
// 	os.Exit(-1)
// }

