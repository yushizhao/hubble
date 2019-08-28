package main

import (
	"github.com/wonderivan/logger"
	"github.com/yushizhao/hubble/config"
	"github.com/yushizhao/hubble/server"
)

//`{
//	"TimeFormat":"2006-01-02 15:04:05", // 输出日志开头时间格式
//	"Console": {            // 控制台日志配置
//		"level": "TRAC",    // 控制台日志输出等级
//			"color": true       // 控制台日志颜色开关
//	},
//	"File": {                   // 文件日志配置
//		"filename": "app.log",  // 初始日志文件名
//			"level": "TRAC",        // 日志文件日志输出等级
//			"daily": true,          // 跨天后是否创建新日志文件，当append=true时有效
//			"maxlines": 1000000,    // 日志文件最大行数，当append=true时有效
//			"maxsize": 1,           // 日志文件最大大小，当append=true时有效
//			"maxdays": -1,          // 日志文件有效期
//			"append": true,         // 是否支持日志追加
//			"permit": "0660"        // 新创建的日志文件权限属性
//	},
//	"Conn": {                       // 网络日志配置
//		"net":"tcp",                // 日志传输模式
//			"addr":"10.1.55.10:1024",   // 日志接收服务器
//			"level": "Warn",            // 网络日志输出等级
//			"reconnect":true,           // 网络断开后是否重连
//			"reconnectOnMsg":false,     // 发送完每条消息后是否断开网络
//	}
//}`

var (
	logCfg = `{
	"TimeFormat":"2006-01-02 15:04:05", 
	"Console": {           
		"level": "DEBG",  
		"color": true       
	},
	"File": {                   
		"filename": "hubble.log", 
			"level": "INFO",        
			"daily": true,         
			"maxlines": 1000000,   
			"maxsize": 10,          
			"maxdays": -1,          
			"append": true,        
			"permit": "0660"        
	}
}`
)

//等级	配置	释义	控制台颜色
//0	EMER	系统级紧急，比如磁盘出错，内存异常，网络不可用等	红色底
//1	ALRT	系统级警告，比如数据库访问异常，配置文件出错等	紫色
//2	CRIT	系统级危险，比如权限出错，访问异常等	蓝色
//3	EROR	用户级错误	红色
//4	WARN	用户级警告	黄色
//5	INFO	用户级重要	天蓝色
//6	DEBG	用户级调试	绿色
//7	TRAC	用户级基本输出，比如成员信息，结构体值等	绿色

func main() {
	//config.GlobalInstance()//读配置，初始化全局变量
	// 通过配置参数直接配置
	logger.SetLogger(logCfg)

	err := config.ReadConfig()
	if err != nil {
		logger.Error(err)
	}

	if config.Server.MOCK {
		server.MOCK_StartServer()
	} else {
		server.StartServer()
	}

	done := make(chan bool, 1)
	<-done
}
