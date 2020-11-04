package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage = `mydocker is a simple container runtime implementation.
			   The purpose of this project is to learn how docker works and how to write a docker by ourselves
			   Enjoy it, just for fun.`

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	//定义命令行
	app.Commands = []cli.Command{
		initCommand,   //init命令（容器内部调用）
		runCommand,  //run命令
	}

	//日志
	app.Before = func(context *cli.Context) error {
		// Log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.JSONFormatter{})

		log.SetOutput(os.Stdout)
		return nil
	}

	//启动程序，并解析命令行，进行相应调用-------->main_command.go
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

/*
go build .      编译后：
./mydocker run -ti /bin/sh    其中－ti 表示想要以交互式的形式运行容器，/bin/sh 为指定容器内运行的第一个进程
ps -ef
./mydocker run -ti /bin/ls


第1次启动后，再启动，报错：
fork/exec /proc/self/exe: no such file or directory
https://blog.csdn.net/qq_27068845/article/details/90705925
怎么恢复？？？重启系统

*/
