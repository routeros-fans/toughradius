package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	_ "time/tzdata"

	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/config"
	"github.com/talkincode/toughradius/v8/controllers"
	"github.com/talkincode/toughradius/v8/docs"
	"github.com/talkincode/toughradius/v8/freeradius"
	"github.com/talkincode/toughradius/v8/installer"
	"github.com/talkincode/toughradius/v8/toughradius"
	"github.com/talkincode/toughradius/v8/tr069"
	"github.com/talkincode/toughradius/v8/webserver"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group

	BuildVersion   string
	ReleaseVersion string
	BuildTime      string
	BuildName      string
	CommitID       string
	CommitDate     string
	CommitUser     string
	CommitSubject  string
)

// 命令行定义
var (
	h         = flag.Bool("h", false, "help usage")
	showVer   = flag.Bool("v", false, "show version")
	conffile  = flag.String("c", "", "config yaml file")
	initdb    = flag.Bool("initdb", false, "run initdb")
	install   = flag.Bool("install", false, "run install")
	uninstall = flag.Bool("uninstall", false, "run uninstall")
	initcfg   = flag.Bool("initcfg", false, "write default config > /etc/toughradius.yml")
)

// PrintVersion Print version information
func PrintVersion() {
	_, _ = fmt.Fprintf(os.Stdout, "build name:\t%s\n", BuildName)
	_, _ = fmt.Fprintf(os.Stdout, "build version:\t%s\n", BuildVersion)
	_, _ = fmt.Fprintf(os.Stdout, "build time:\t%s\n", BuildTime)
	_, _ = fmt.Fprintf(os.Stdout, "release version:\t%s\n", ReleaseVersion)
	_, _ = fmt.Fprintf(os.Stdout, "Commit ID:\t%s\n", CommitID)
	_, _ = fmt.Fprintf(os.Stdout, "Commit Date:\t%s\n", CommitDate)
	_, _ = fmt.Fprintf(os.Stdout, "Commit Username:\t%s\n", CommitUser)
	_, _ = fmt.Fprintf(os.Stdout, "Commit Subject:\t%s\n", CommitSubject)
}

func printHelp() {
	if *h {
		ustr := fmt.Sprintf("%s version: %s, Usage:%s -h\nOptions:", BuildName, BuildVersion, BuildName)
		_, _ = fmt.Fprintf(os.Stderr, ustr)
		flag.PrintDefaults()
		os.Exit(0)
	}
}

//	@title			Toughradius API
//	@version		1.0
//	@description	This is Toughradius API
//	@termsOfService	https://github.com/talkincode/toughradius
//	@contact.name	Toughradius API Support
//	@contact.url	https://github.com/talkincode/toughradius
//	@contact.email	jamiesun.net@gmail.com
//	@license.name	GPL
//	@license.url	https://github.com/talkincode/toughradius

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Accesskey based security scheme to secure api

// @host		127.0.0.1:1816
// @BasePath	/
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if *showVer {
		PrintVersion()
		os.Exit(0)
	}

	printHelp()

	_config := config.LoadConfig(*conffile)
	if *initcfg {
		err := installer.InitConfig(_config)
		if err != nil {
			log.Error(err)
		}
		return
	}

	// Install as a system service
	if *install {
		err := installer.Install()
		if err != nil {
			log.Error(err)
		}
		return
	}

	// 卸载
	if *uninstall {
		installer.Uninstall()
		return
	}

	if *initdb {
		app.InitGlobalApplication(_config)
		app.GApp().InitDb()
		return
	}

	app.InitGlobalApplication(_config)
	docs.Update()

	app.GApp().MigrateDB(false)

	defer app.Release()

	// 管理服务启动
	g.Go(func() error {
		webserver.Init()
		controllers.Init()
		return webserver.Listen()
	})

	// The Freeradius API service starts
	g.Go(func() error {
		return freeradius.Listen()
	})

	radiusService := toughradius.NewRadiusService()
	defer radiusService.Release()

	g.Go(func() error {
		return toughradius.ListenRadiusAuthServer(toughradius.NewAuthService(radiusService))
	})

	g.Go(func() error {
		return toughradius.ListenRadiusAcctServer(toughradius.NewAcctService(radiusService))
	})

	g.Go(func() error {
		radsec := toughradius.NewRadsecService(
			toughradius.NewAuthService(radiusService),
			toughradius.NewAcctService(radiusService),
		)
		return toughradius.ListenRadsecServer(radsec)
	})

	g.Go(func() error {
		log.Info("Start tr069 server...")
		return tr069.Listen()
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
