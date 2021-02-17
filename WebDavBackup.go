package main

import (
	"flag"
	"fmt"
	logging "github.com/op/go-logging"
	wd "github.com/studio-b12/gowebdav"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	//"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Conf struct {
	Duration int
	LogLvl   string     `yaml:"logLvl"`
	TasksWD  []ConfTask `yaml:"Tasks"`
}

type ConfTask struct {
	Host          string `yaml:"Host"`
	User          string
	Password      string
	RetryAttempts int    `yaml:"retryAttempts"`
	ArcDir        string `yaml:"arc_dir"`
	TzCorrection  string `yaml:"TZCorrection"`
	LocDir        string `yaml:"loc_dir"`
	FileName      string `yaml:"file_name"`
}

func (c ConfTask) String() string {
	return fmt.Sprintf("{Host:%s User:%s Password:******** ArcDir:%s TzCorrection:%s LocDir:%s FileName:%s}",
		c.Host, c.User, c.ArcDir, c.TzCorrection, c.LocDir, c.FileName)
}

func ReadConfig(path string) (*Conf, error) {
	confFile, err := os.Open(path)
	if err != nil {
		configLog.Error(err)
		//log.Printf("err ReadConfig\t%s", err)
		return nil, err
	}
	//noinspection GoUnhandledErrorResult
	defer confFile.Close()
	if confFile != nil {
		fileAll, _ := ioutil.ReadAll(confFile)
		//log.Println(string(fileAll))
		configLog.Info("################# READING CONFIGURATION #################")
		//log.Println("################# READING CONFIGURATION #################")
		currConf := Conf{}

		err := yaml.Unmarshal(fileAll, &currConf)
		if err != nil {
			configLog.Critical(err)
			//log.Fatalf("error: %v", err)
			return nil, err
		}
		configLog.Infof("Applyed config:\n%+v\n", currConf)
		//log.Printf("Applyed config:\n%+v\n", currConf)
		configLog.Info("#########################################################")
		//log.Print("#########################################################")
		return &currConf, nil
	}
	return nil, fmt.Errorf("file %s is empty", path)
}

func ParseArcDate(arcName string, confName string) (time.Time, error) {
	layout := "02-01-2006_15.04.05"
	layoutFull := "02-01-2006_15.04.05-0700"
	extConfName := filepath.Ext(confName)
	prefixConfName := confName[:len(confName)-len(extConfName)] + "_"
	//log.Printf("extConfName: %s; prefixConfName: %s", extConfName, prefixConfName)
	needLen := len(layout) + len(extConfName) + len(prefixConfName)
	needLenFull := len(layoutFull) + len(extConfName) + len(prefixConfName)
	//log.Printf("extConfName: %s; prefixConfName: %s; needLen: %d; needLenFull: %d; lenArcName: %d", extConfName,
	//	prefixConfName,	needLen, needLenFull, len(arcName))
	if (len(arcName) != needLen) && (len(arcName) != needLenFull) {
		return time.Time{}, fmt.Errorf("the length of the file name \"%s\" differs from the required length",
			arcName)
	}
	hasPref := strings.HasPrefix(arcName, prefixConfName)
	if hasPref != true {
		return time.Time{}, fmt.Errorf("NOT has prefix \"%s\" in name file \"%s\" from arhive directory",
			prefixConfName, arcName)
	}
	arcNameWithoutPref := strings.TrimPrefix(arcName, prefixConfName)
	hasSuf := strings.HasSuffix(arcName, extConfName)
	if hasSuf != true {
		return time.Time{}, fmt.Errorf("NOT has extention \"%s\" in name file \"%s\" from arhive directory",
			extConfName, arcName)
	}
	strArcDate := strings.TrimSuffix(arcNameWithoutPref, extConfName)
	//log.Printf("Total date from file name arc file: %s", strArcDate)
	if len(strArcDate) == len(layoutFull) {
		parseDate, err := time.Parse(layoutFull, strArcDate)
		if err != nil {
			return time.Time{}, err
		}
		return parseDate, nil
	} else {
		parseDate, err := time.ParseInLocation(layout, strArcDate, time.Local)
		if err != nil {
			return time.Time{}, err
		}
		return parseDate, nil
	}
	//return time.Time{}, nil
}

var configLog = logging.MustGetLogger("forConfig")
var log = logging.MustGetLogger("mainLog")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
// Def format: `%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`
var formatToConfig = logging.MustStringFormatter(
	`%{color}%{time:2006/01/02 15:04:05.000}  %{shortfunc} CONF%{color:reset} %{message}`,
)
var format = logging.MustStringFormatter(
	`%{color}%{time:2006/01/02 15:04:05.000}  %{shortfunc} %{level:.4s}%{color:reset} %{message}`,
)

func main() {
	confBackend := logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(os.Stdout, "", 0), formatToConfig))
	//	"CRITICAL", "ERROR", "WARNING", "NOTICE", "INFO", "DEBUG"
	confBackend.SetLevel(logging.INFO, "")
	configLog.SetBackend(confBackend)

	backend := logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(os.Stdout, "", 0), format))
	//backend.SetLevel(logging.INFO, "")
	log.SetBackend(backend)

	//log.Info("info")
	//log.Notice("notice")
	//log.Warning("warning")
	//log.Error("err")
	//log.Critical("crit")

	var confPath string

	//parsing flag "--conf"
	flag.StringVar(&confPath, "conf", "config.yml", "Path to config")
	flag.Parse()
	configLog.Infof("path to config: %s", confPath)

	conf, err := ReadConfig(confPath)
	if err != nil {
		configLog.Critical(err)
		return
	}

	CheckConfig(conf)

	durSleep := time.Duration(conf.Duration) * time.Minute

	var logLvl logging.Level
	//maps config log level to library level (github)
	switch strings.ToUpper(conf.LogLvl) {
	case "DEBUG":
		logLvl = logging.DEBUG
	case "INFO":
		logLvl = logging.INFO
	case "WORN":
		logLvl = logging.WARNING
	case "ERR":
		logLvl = logging.ERROR
	default:
		logLvl = logging.INFO
		log.Warningf("the value of parameter \"logLvl\" in the configuration file is set incorrectly "+
			"(\"%s\")", conf.LogLvl)
		log.Warningf("the default value was applied. logLvl: %s", logLvl)
	}

	backend.SetLevel(logLvl, "")

	lastModTimeMap := make(map[string]time.Time)

	for {
		//Main cycle. Listing set directories(tasks) from config and do backup
		for _, currTask := range conf.TasksWD {
			pathToBackup := currTask.LocDir + currTask.FileName
			lastModTime := lastModTimeMap[pathToBackup]
			log.Debugf("(%s)\tlastModTime: %s", pathToBackup, lastModTime)
			modTimeLocalFile, err := getLocalModTime(pathToBackup)
			if err != nil {
				log.Errorf("(%s)\t%s", currTask.LocDir, err)
			}
			log.Infof("file: %s, modTimeLocalFile: %v (to cache: %v)", currTask.FileName,
				modTimeLocalFile.Format("02-01-2006 15.04.05 -0700 MST"),
				lastModTime.Format("02-01-2006 15.04.05 -0700 MST"))
			if modTimeLocalFile.Unix() > lastModTime.Unix() {
				err = doBackup(currTask, modTimeLocalFile)
				if err != nil {
					log.Errorf("(%s)\t%s", currTask.LocDir, err)
				} else {
					lastModTimeMap[pathToBackup] = modTimeLocalFile
					log.Debugf("(%s)\tNEW lastModTime: %s", currTask.LocDir, lastModTimeMap[pathToBackup])
				}
			}
		}
		log.Infof("Waiting %v minuts, next running at %v", durSleep.Minutes(),
			time.Now().Add(durSleep).Format("02-01-2006 15.04.05 -0700 MST"))
		time.Sleep(durSleep)
	}

}

func CheckConfig(conf *Conf) {
	for _, task := range conf.TasksWD {
		if task.RetryAttempts < 1 {
			log.Panicf("the value of parameter \"retryAttempts\" from Task(%s) in the configuration file "+
				"is set incorrectly (\"%v\"). THE APP IS STOPPED!", task.LocDir+task.FileName, task.RetryAttempts)
		}
	}
}

func GetArcLastDate(currTask ConfTask, wdServer *wd.Client) (time.Time, error) {
	arcDir := currTask.ArcDir
	var filesArch []os.FileInfo
	var err error
	//Get all in archive directory on WebDav
	for i := 0; i < currTask.RetryAttempts; i++ {
		filesArch, err = wdServer.ReadDir(arcDir)
		if err == nil {
			break
		}
	}
	if err != nil {
		return time.Time{}, err
	}
	backupFileName := currTask.FileName
	return DoGetArcLastDate(backupFileName, filesArch)
}

func DoGetArcLastDate(backupFileName string, filesArch []os.FileInfo) (time.Time, error) {
	curExt := filepath.Ext(backupFileName)
	var arcLastDate time.Time = time.Unix(0, 0) //nil
	log.Debugf("Serching files on archive directory with the extension: %s", curExt)
	for _, file := range filesArch {
		//Only files with the required extension
		if (filepath.Ext(file.Name()) == curExt) && (file.IsDir() == false) {
			//log.Println(file)
			//fileModeTime := file.ModTime()
			fileParseModeTime, err := ParseArcDate(file.Name(), backupFileName)
			if err != nil {
				log.Warning(err)
				continue
			}
			//log.Printf("Testing parse date from file name on archive:  parse date: %s, filename: %s",
			//	fileParseModeTime, file.Name())
			//Used filemode time from atribute
			//if fileModeTime.Unix() > arcLastDate.Unix() {
			//	arcLastDate = fileModeTime
			//}
			//log.Println(file.Name())

			//Used get filemode from file name
			if fileParseModeTime.Unix() > arcLastDate.Unix() {
				arcLastDate = fileParseModeTime
			}
		}
	}
	return arcLastDate, nil
}

func doBackup(currTask ConfTask, modTimeLocalFile time.Time) error {
	pathToBackup := path.Join(currTask.LocDir, currTask.FileName)
	log.Infof("Current path to backup file: %s; path to archive: %s",
		pathToBackup, currTask.Host+currTask.ArcDir)
	//log.Printf("inf doBackup\tCurrent path to backup file: %s; path to archive: %s",
	//	pathToBackup, currTask.Host+currTask.ArcDir)

	log.Debugf("Trying auth to WebDav server: %s, user: %s\n",
		currTask.Host, currTask.User)
	wdServer := wd.NewClient(currTask.Host, currTask.User, currTask.Password)
	//arcDir := currTask.ArcDir

	//Find Arc dir on WebDav, create it if not found
	err := CreateRemoteArcDirIfNotExists(currTask, wdServer)
	if err != nil {
		return err
	}

	err = uploadFileIfNeeded(currTask, wdServer, modTimeLocalFile)
	if err != nil {
		return err
	}
	return nil
}

func getLocalModTime(pathToBackup string) (time.Time, error) {
	statFileBackup, err := os.Stat(pathToBackup)
	if err != nil {
		return time.Time{}, err
	}
	modTimeLocalFile := statFileBackup.ModTime()
	log.Debugf("Backup file info UNIX: %d, Time: %v, Location: %s",
		modTimeLocalFile.Unix(), modTimeLocalFile, modTimeLocalFile.Location())
	return modTimeLocalFile, nil
}

func uploadFileIfNeeded(currTask ConfTask, wdServer *wd.Client, modTimeLocalFile time.Time) error {
	pathToBackup := path.Join(currTask.LocDir, currTask.FileName)
	arcDir := currTask.ArcDir

	//Searching last file and max date on archive directory
	arcLastDate, err := GetArcLastDate(currTask, wdServer)
	if err != nil {
		return err
	}
	//Coerce the last date of the backup file from WebDav archive to a local TimeZone
	arcLastDate = arcLastDate.In(modTimeLocalFile.Location())
	log.Infof("Actual arcLastDate in local TimeZone: %s", arcLastDate)
	//Making new filename to backup, if current file is newest
	if modTimeLocalFile.Unix() > arcLastDate.Unix() {
		newFileName, err := CreateNewFileName(currTask, modTimeLocalFile)
		if err != nil {
			return err
		}
		bytes, _ := ioutil.ReadFile(pathToBackup)
		pathToNewBackupFile := path.Join(arcDir, newFileName)
		log.Infof("Uploading backup file to WebDev path: %s", pathToNewBackupFile)
		for i := 0; i < currTask.RetryAttempts; i++ {
			err = wdServer.Write(pathToNewBackupFile, bytes, 0644)
			if err == nil {
				break
			}
		}
		if err != nil {
			return err
		}
		log.Info("Copy successed")
	}
	return nil
}

func CreateNewFileName(currTask ConfTask, modTimeLocalFile time.Time) (string, error) {
	var correctTzString string
	tzConfStr := currTask.TzCorrection
	re := regexp.MustCompile("[0-9]+")
	digits := re.FindAllString(tzConfStr, -1)
	if strings.HasPrefix(tzConfStr, "-") {
		correctTzString = "-" + strings.Join(digits, "")
	} else {
		correctTzString = "+" + strings.Join(digits, "")
	}
	timeToTzOffset, err := time.Parse("-0700", correctTzString)
	if err != nil {
		return "", nil
	}
	_, offsetConf := timeToTzOffset.Zone()
	_, offsetLocalFile := modTimeLocalFile.Zone()
	if offsetConf != offsetLocalFile {
		modTimeLocalFile = modTimeLocalFile.In(timeToTzOffset.Location())
	}
	curExt := filepath.Ext(currTask.FileName)
	timeToName := modTimeLocalFile.Format("02-01-2006_15.04.05-0700")
	nameWithoutExt := currTask.FileName[:len(currTask.FileName)-len(curExt)]
	newFileName := nameWithoutExt + "_" + timeToName + curExt
	log.Infof("Name without extention: %s", nameWithoutExt)
	log.Infof("New Filename to backup file: %s", newFileName)
	return newFileName, nil
}

func CreateRemoteArcDirIfNotExists(currTask ConfTask, wdServer *wd.Client) error {
	arcDir := currTask.ArcDir
	var filesRoot []os.FileInfo
	var err error
	//retry if got error
	for i := 0; i < currTask.RetryAttempts; i++ {
		//Get all files and directories on WebDav root
		filesRoot, err = wdServer.ReadDir("/")
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	//Count item on WebDav root
	cnt := len(filesRoot)
	log.Debugf("Count item on WebDav root directory: %d", cnt)
	//Listing filesRoot and directories, find arch directory
	for _, file := range filesRoot {
		cnt -= 1
		if (file.Name() == arcDir) && (file.IsDir() == true) {
			log.Debugf("Archive directory \"%s\" found", arcDir)
			break
		} else if cnt == 0 {
			log.Warningf("Archive directory \"%s\" NOT found", arcDir)
			//Join root WebDav path and archive directory
			arcUrl, err := url.Parse(currTask.Host)
			if err != nil {
				return err
			}
			arcUrl.Path = path.Join(arcUrl.Path, arcDir)
			arcUrlStr := arcUrl.String()
			log.Infof("Creating archive directory on path: %s", arcUrlStr)
			for i := 0; i < currTask.RetryAttempts; i++ {
				err = wdServer.Mkdir(arcDir, 700)
				if err == nil {
					break
				}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
