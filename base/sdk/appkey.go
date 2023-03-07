package sdk

type AppInfo struct {
	Secret   string
	Platform string // android/ios
	AppId    int    // 马甲包ID
}

// {APP_KEY: {APP_SECRET,...}}
var appKeyStore map[string]*AppInfo
var sysKeyStore map[string]string

func init() {
	appKeyStore = make(map[string]*AppInfo)
	// android APP:1
	appKeyStore["5vukw46u8d"] = &AppInfo{Secret: "hqh6omd91vxmknx0zn2jcy7ddnsncyqa", Platform: "android", AppId: 1}
	// ios APP:1
	appKeyStore["8q77qtxdo6"] = &AppInfo{Secret: "9qq822wcmejar1rpc6g6irsg5i83x3nl", Platform: "ios", AppId: 1}
	// web APP:1 zt5kqoocul3l1otvnjxoltbpx4vnteyn
	appKeyStore["aei2z9o2a6"] = &AppInfo{Secret: "", Platform: "web", AppId: 1}
	// desktop APP:1
	appKeyStore["k6q7w8dcvl"] = &AppInfo{Secret: "878ujst19q41auwlg4k4p9siv8luvmhc", Platform: "desktop", AppId: 1}
	// sys
	appKeyStore["oidr3ty0im"] = &AppInfo{Secret: "qp23fue6pk4mop3agt48bjoz0drpejoq", Platform: "sys", AppId: 1}

	sysKeyStore = make(map[string]string)
	// admin
	sysKeyStore["oidr3ty0im"] = "qp23fue6pk4mop3agt48bjoz0drpejoq"
}

func GetAppInfo(key string) *AppInfo {
	return appKeyStore[key]
}

func GetSysSecret(key string) string {
	s, _ := sysKeyStore[key]
	return s
}
