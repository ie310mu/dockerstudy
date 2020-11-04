package main

import (
	"dockerstudy/5.1/mydocker/container"
	"dockerstudy/5.1/mydocker/cgroups/subsystems"
	"dockerstudy/5.1/mydocker/cgroups"
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"
)


func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	// use mydocker-cgroup as cgroup name
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	//后台运行时容器未退出,cgroupManager.Destroy会报错 
	//remove cgroup fail remove /sys/fs/cgroup/memory/mydocker-cgroup/cgroup.procs: operation not permitted
	defer cgroupManager.Destroy()       
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	sendInitCommand(comArray, writePipe)
	if tty {
		parent.Wait()
	}
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
