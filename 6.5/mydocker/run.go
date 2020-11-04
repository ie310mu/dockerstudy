package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"dockerstudy/6.5/mydocker/cgroups"
	"dockerstudy/6.5/mydocker/cgroups/subsystems"
	"dockerstudy/6.5/mydocker/container"
	"dockerstudy/6.5/mydocker/network"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"os/exec"
	"time"
)

//go 调用 linux shell测试
//注意，不能在容器中调用，会报错   ls / error open /dev/null: no such file or directory
func testRunShell(){
	//方式1
	if data, err := exec.Command("/bin/ls", "/").CombinedOutput(); err != nil {
		log.Errorf("ls / error %v", err)
	}else{
		msg := string(data)
		// log.Info(msg)
		fmt.Println(msg)
	}

	
	fmt.Println("")
	fmt.Println("")

	//方式2
	path, err2 := exec.LookPath("ls")
	if err2 != nil {
		log.Errorf("Exec loop path error %v", err2)
	}
	log.Infof("Find path %s", path)
	cmd := exec.Command(path, "/")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err2 = cmd.Run()
	if err2 != nil {
		fmt.Println("Execute Command failed:" + err2.Error())
		return
	}
}

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, containerName, volume, imageName string,
	envSlice []string, nw string, portmapping []string) {	
	// testRunShell();
		
	containerID := randStringBytes(10)
	if containerName == "" {
		containerName = containerID
	}

	parent, writePipe := container.NewParentProcess(tty, containerName, volume, imageName, envSlice)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}

	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	//record container info
	containerName, err := recordContainerInfo(parent.Process.Pid, comArray, containerName, containerID, volume)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}

	// use containerID as cgroup name
	cgroupManager := cgroups.NewCgroupManager(containerID)
	//后台运行时容器未退出,cgroupManager.Destroy会报错 
	//remove cgroup fail remove /sys/fs/cgroup/memory/mydocker-cgroup/cgroup.procs: operation not permitted
	if tty {
		defer cgroupManager.Destroy()
	}
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	if nw != "" {
		// config container network
		network.Init()
		containerInfo := &container.ContainerInfo{
			Id:          containerID,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portmapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			log.Errorf("Error Connect Network %v", err)
			return
		}
	}

	sendInitCommand(comArray, writePipe)

	if tty {
		parent.Wait()
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume, containerName)
	}

}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func recordContainerInfo(containerPID int, commandArray []string, containerName, id, volume string) (string, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	containerInfo := &container.ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
