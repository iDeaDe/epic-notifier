package app

import (
	"flag"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type NotifierConfig struct {
	Channel                 string
	AnnounceRecheckInterval int
	RemindPostDelay         int
}

type Params struct {
	// Тихая отправка постов(без уведомления подписчикам)
	Silent bool

	// Параметры, влияющие на переотправку постов
	Current bool
	Next    bool
	Remind  bool

	// Технические аргументы
	onlyInit bool
}

type App struct {
	WorkDir string
	Config  NotifierConfig
	Params  Params
	viper   *viper.Viper
}

var defaultParams *Params
var logger *zerolog.Logger

func init() {
	defaultParams = parseFlags()
}

func detectWorkdir() string {
	resultWorkdir := os.Getenv("WORKDIR")

	if resultWorkdir == "" {
		var err error
		executable, err := os.Executable()
		if err != nil {
			resultWorkdir = "."
		} else {
			resultWorkdir = filepath.Dir(executable)
		}
	}

	return filepath.Clean(resultWorkdir)
}

func parseFlags() *Params {
	var params Params

	flag.BoolVar(&params.Silent, "s", false, "Post games silently.")
	flag.BoolVar(&params.Current, "c", false, "Post current games.")
	flag.BoolVar(&params.Next, "n", false, "Create new post with games of the next giveaway.")
	flag.BoolVar(&params.Remind, "r", false, "Resend remind post to the channel.")
	flag.BoolVar(&params.onlyInit, "init", false, "Only initialize app, but not run.")
	flag.Parse()

	return &params
}

func Logger() *zerolog.Logger {
	if logger == nil {
		instance := zerolog.New(zerolog.NewConsoleWriter())
		logger = &instance
	}

	return logger
}

func InitApp() (*App, error) {
	app := new(App)
	app.WorkDir = detectWorkdir()
	app.Params = *defaultParams

	err := app.initConfig()
	if err != nil {
		Logger().Panic().Err(err).Send()
	}

	if app.Params.onlyInit {
		Logger().Info().Msg("app initialized successfully")
		os.Exit(0)
	}

	return app, nil
}

func (a *App) initConfig() error {
	a.viper = viper.New()

	a.viper.AddConfigPath(a.WorkDir)
	a.viper.SetConfigName("config")
	a.viper.SetConfigType("toml")
	a.viper.WatchConfig()

	err := a.viper.SafeWriteConfig()
	if err != nil {
		return err
	}

	if err := a.viper.ReadInConfig(); err != nil {
		return err
	}

	a.viper.
		OnConfigChange(func(in fsnotify.Event) {
			if in.Has(fsnotify.Write) {
				err := a.viper.Unmarshal(&a.Config)
				if err != nil {
					Logger().Panic().Err(err).Send()
				}
			}
		})

	err = a.viper.Unmarshal(&a.Config)
	if err != nil {
		return err
	}

	return nil
}
