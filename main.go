package main

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ivartj/kartoteka/controller"
	"github.com/ivartj/kartoteka/core"
	"github.com/ivartj/kartoteka/repository"
	"github.com/ivartj/minn/args"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	mainProgramName    = "kartoteka"
	mainProgramVersion = "0.1-SNAPSHOT"
)

type mainConfiguration struct {
	Port            uint16
	Database        string
	AssetsDirectory string
	DefaultLanguage language.Tag
}

var defaultConfiguration = mainConfiguration{
	Port:            8888,
	Database:        "./kartoteka.db",
	AssetsDirectory: "./assets",
	DefaultLanguage: language.English,
}

func mainUsage(out io.Writer) {
	fmt.Fprintf(out, "Usage: %s [ -p PORT ]\n", mainProgramName)
}

func mainParseArgs(argv []string, cfg *mainConfiguration, log core.Logger) error {

	tok := args.NewTokenizer(argv)

	for tok.Next() {
		switch tok.Arg() {

		case "-h", "-?", "--help":
			mainUsage(os.Stdout)
			os.Exit(0)

		case "--version":
			fmt.Printf("%s version %s\n", mainProgramName, mainProgramVersion)
			os.Exit(0)

		case "-p", "--port":
			portStr, err := tok.TakeParameter()
			if err != nil {
				return err
			}
			port, err := strconv.ParseUint(portStr, 10, 16)
			if err != nil {
				return err
			}
			cfg.Port = uint16(port)

		case "--database":
			var err error
			cfg.Database, err = tok.TakeParameter()
			if err != nil {
				return err
			}

		case "--assets-directory":
			var err error
			cfg.AssetsDirectory, err = tok.TakeParameter()
			if err != nil {
				return err
			}

		case "--default-language":
			languageTag, err := tok.TakeParameter()
			if err != nil {
				return err
			}
			cfg.DefaultLanguage, err = language.Parse(languageTag)
			if err != nil {
				return fmt.Errorf("Failed parsing language tag: %w", err)
			}

		default:
			log.Fatalf("Unrecognized option, '%s'", tok.Arg())
		}
	}

	if tok.Err() != nil {
		return tok.Err()
	}

	return nil
}

func mainLoadI18n(i18nDirectory string, defaultLanguage language.Tag) (*i18n.Bundle, error) {
	bundle := i18n.NewBundle(defaultLanguage)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	dirpath := i18nDirectory
	dir, err := os.Open(dirpath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open i18n (%s) directory: %w", dirpath, err)
	}
	dirnames, err := dir.Readdirnames(0)
	if err != nil {
		return nil, fmt.Errorf("Failed to read i18n (%s) directory: %w", dirpath, err)
	}
	tomlnames := make([]string, 0, len(dirnames))
	for _, name := range dirnames {
		if strings.HasSuffix(name, ".toml") {
			tomlnames = append(tomlnames, name)
		}
	}
	for _, filename := range tomlnames {
		filepath := dirpath + "/" + filename
		_, err = bundle.LoadMessageFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("Failed to load %s: %w", filepath, err)
		}
	}
	return bundle, nil
}

func mainOpenDatabase(filename string) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		x := recover()
		if x != nil || err != nil {
			db.Close()
		}
	}()

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = repository.InitSchema(tx)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func mainParseTemplateFiles(templateDirectory string) (*template.Template, error) {
	dirFile, err := os.Open(templateDirectory)
	if err != nil {
		return nil, err
	}
	fis, err := dirFile.Readdir(0)
	if err != nil {
		return nil, err
	}
	templatePaths := []string{}
	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".template.html") {
			templatePaths = append(templatePaths, path.Join(templateDirectory, fi.Name()))
		}
	}
	tpl := template.New("main")
	tpl.Funcs(template.FuncMap(map[string]interface{}{
		"tr": func(localizer *i18n.Localizer, messageID, defaultMessage string) (string, error) {
			return localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    messageID,
					Other: defaultMessage,
				},
			})
		},
	}))
	_, err = tpl.ParseFiles(templatePaths...)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func mainHTTPHandler(db *sql.DB, tpl *template.Template, i18nBundle *i18n.Bundle, staticDirectory string) http.Handler {
	mux := http.NewServeMux()

	random := controller.NewRandom(db, tpl, i18nBundle)
	mux.Handle("/random", random)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDirectory))))

	return mux
}

func main() {
	logger := log.New(os.Stderr, "kartoteka: ", 0)

	cfg := defaultConfiguration
	err := mainParseArgs(os.Args, &cfg, logger)
	if err != nil {
		logger.Fatalf("Error on parsing command line arguments: %s", err)
	}

	i18nBundle, err := mainLoadI18n(cfg.AssetsDirectory+"/i18n", cfg.DefaultLanguage)
	if err != nil {
		logger.Fatalf("Failed to load localization messages: %s", err)
	}

	db, err := mainOpenDatabase(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to open database file: %s", err)
	}

	tpl, err := mainParseTemplateFiles(cfg.AssetsDirectory + "/templates")
	if err != nil {
		logger.Fatalf("Error parsing template files: %s", err)
	}

	handler := mainHTTPHandler(db, tpl, i18nBundle, cfg.AssetsDirectory+"/static")
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler)
	if err != nil {
		logger.Fatalf("Error serving HTTP requests: %s", err)
	}
}
