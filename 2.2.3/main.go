package main

import (
	"os/exec"
	"path"
	"os"
	"fmt"
	"io/ioutil"
	"syscall"
	"strconv"
)

//挂载了 memory subsystem hierarchy 的根目录位置
const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"

func main(){
	//===========fork出来的进程的操作============
	if os.Args[0] == "/proc/self/exe" {  //  /proc/self/exe代表当前程序（后面代码中为fork出来的当前程序的新进程）  http://www.360doc.com/content/17/0122/09/33093582_624104827.shtml
		//容器进程
		fmt.Printf("current pid  %d", syscall.Getpid())
		fmt.Println()
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)  //在新进程下启动stress，指定200m内存  stress说明：https://www.cnblogs.com/sparkdev/p/10354947.html
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin 	
		cmd.Stdout = os.Stdout 
		cmd.Stderr = os.Stderr 
		if err:= cmd.Run() ; err!= nil {  //在宿主机上用  ps aux | grep stress  查看信息  https://blog.csdn.net/weixin_34029680/article/details/91447939
			fmt.Println(err)
			os.Exit(1)
		}
	}
	//===========fork出来的进程的操作============

    //下面是程序在宿主机上启动的流程
	cmd := exec.Command("/proc/self/exe")  //指定被 fork 来的新进程内的初始命令,默认使用sh来执行
	cmd.SysProcAttr = &syscall.SysProcAttr{ 
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,  
	}
	//syscall.CLONE_NEWUSER调用报错：https://github.com/xianlubird/3.1/mydocker/issues/3
	//centos默认的没有开启user namespace，参考链接https://zhuanlan.zhihu.com/p/31871814  https://github.com/golang/go/issues/16283
	//如果已经内核开启user_namespace依旧invalid argument的话使用下面的命令fixed  echo 640 > /proc/sys/user/max_user_namespaces  https://unix.stackexchange.com/questions/479635/unable-to-create-user-namespace-in-rhel?rq=1
	//最终是注释下面这句:
	//cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(1 ), Gid : uint32 (1)}

	cmd.Stdin = os.Stdin 	
	cmd.Stdout = os.Stdout 
	cmd.Stderr = os.Stderr 

	if err:= cmd.Start() ; err != nil {  //注意cmd.Run和Start的区别  https://blog.csdn.net/zistxym/article/details/8672927
		fmt.Println("ERROR", err) 
		os.Exit(1)
	}else{
		//得到 fork 出来进程映射在外部命名空间的 pid
		fmt.Printf("%v", cmd.Process.Pid)

		//在系统默认创建挂载的 memory subsystem Hierarchy 上创建 cgroup
		os.Mkdir(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit"), 0755)
		//将容器进程加入到这个 cgroup中
		ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)  //*******这句很重要
		//限制 cgroup 进程使用
		ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "memory.limit_in_bytes") , []byte ("100m"), 0644)
	}
	cmd.Process.Wait()	
}