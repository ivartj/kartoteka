package main

import (
	"database/sql"
	"fmt"
	"github.com/ivartj/kartotek/controller"
	"github.com/ivartj/kartotek/core"
	"github.com/ivartj/kartotek/repository"
	"github.com/ivartj/minn/args"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nicksnyder/go-i18n/i18n"
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
	mainProgramName    = "kartotek"
	mainProgramVersion = "0.1-SNAPSHOT"
)

type mainConfiguration struct {
	Port            uint16
	Database        string
	AssetsDirectory string
	DefaultLanguage string
}

var defaultConfiguration = mainConfiguration{
	Port:            8888,
	Database:        "./kartotek.db",
	AssetsDirectory: "./assets",
	DefaultLanguage: "pl",
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
			var err error
			cfg.DefaultLanguage, err = tok.TakeParameter()
			if err != nil {
				return err
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

func mainLoadI18n(i18nDirectory string, log core.Logger) {
	dirpath := i18nDirectory
	dir, err := os.Open(dirpath)
	if err != nil {
		log.Fatalf("Failed to open i18n (%s) directory: %s", dirpath, err.Error())
	}
	dirnames, err := dir.Readdirnames(0)
	if err != nil {
		log.Fatalf("Failed to read i18n (%s) directory: %s", dirpath, err.Error())
	}
	jsonnames := make([]string, 0, len(dirnames))
	for _, name := range dirnames {
		if strings.HasSuffix(name, ".json") {
			jsonnames = append(jsonnames, name)
		}
	}
	for _, filename := range jsonnames {
		filepath := dirpath + "/" + filename
		err := i18n.LoadTranslationFile(filepath)
		if err != nil {
			log.Fatalf("Failed to load %s: %s", filepath, err.Error())
		}
	}
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
	tpl, err := template.New("main").ParseFiles(templatePaths...)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func mainHTTPHandler(db *sql.DB, tpl *template.Template, staticDirectory string) http.Handler {
	mux := http.NewServeMux()

	random := controller.NewRandom(db, tpl)
	mux.Handle("/random", random)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDirectory))))

	return mux
}

func main() {
	logger := log.New(os.Stderr, "kartotek: ", 0)

	cfg := defaultConfiguration
	err := mainParseArgs(os.Args, &cfg, logger)
	if err != nil {
		logger.Fatalf("Error on parsing command line arguments: %s", err)
	}

	mainLoadI18n(cfg.AssetsDirectory+"/i18n", logger)

	db, err := mainOpenDatabase(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to open database file: %s", err)
	}

	tpl, err := mainParseTemplateFiles(cfg.AssetsDirectory + "/templates")
	if err != nil {
		logger.Fatalf("Error parsing template files: %s", err)
	}

	handler := mainHTTPHandler(db, tpl, cfg.AssetsDirectory+"/static")
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler)
	if err != nil {
		logger.Fatalf("Error serving HTTP requests: %s", err)
	}
}
