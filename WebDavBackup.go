package main

import (
	"fmt"
	wd "github.com/studio-b12/gowebdav"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Conf struct {
	Duration int
	TasksWD  []ConfTask `yaml:"Tasks"`
}

type ConfTask struct {
	Host     string `yaml:"Host"`
	User     string
	Password string
	ArcDir   string `yaml:"arc_dir"`
	LocDir   string `yaml:"loc_dir"`
	FileName string `yaml:"file_name"`
}

func ReadConfig(path string) (*Conf, error) {
	confFile, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if confFile != nil {
		fileAll, _ := ioutil.ReadAll(confFile)
		log.Println(string(fileAll))
		log.Println("#########################################################")
		currConf := Conf{}

		err := yaml.Unmarshal(fileAll, &currConf)
		if err != nil {
			log.Fatalf("error: %v", err)
			return nil, err
		}
		log.Printf("Applyed config:\n%+v\n", currConf)
		log.Print("#########################################################")
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

func main() {
	conf, err := ReadConfig("config.yml")
	durSleep := time.Duration(conf.Duration) * time.Minute
	if err != nil {
		log.Fatal(err)
		return
	}

	for {
		//Main cycle. Listing set directories(tasks) from config and do backup
		for _, currTask := range conf.TasksWD {
			err := doBackup(currTask)
			if err != nil {
				log.Printf("err main\t%s", err)
			}
		}
		time.Sleep(durSleep)
	}

}

func GetArcLastDate(currTask ConfTask, wdServer *wd.Client) (time.Time, error) {
	arcDir := currTask.ArcDir
	//Get all in archive directory on WebDav
	filesArch, err := wdServer.ReadDir(arcDir)
	if err != nil {
		return time.Time{}, err
	}
	backupFileName := currTask.FileName
	curExt := filepath.Ext(backupFileName)
	var arcLastDate time.Time = time.Unix(0, 0) //nil
	log.Printf("Serching files on archive directory with the extension: %s", curExt)
	for _, file := range filesArch {
		//Only files with the required extension
		if (filepath.Ext(file.Name()) == curExt) && (file.IsDir() == false) {
			//log.Println(file)
			//fileModeTime := file.ModTime()
			fileParseModeTime, err := ParseArcDate(file.Name(), backupFileName)
			if err != nil {
				log.Printf("err GetArcLastDate\t%s", err)
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

func doBackup(currTask ConfTask) error {
	pathToBackup := path.Join(currTask.LocDir, currTask.FileName)
	log.Printf("Current path to backup file: %s", pathToBackup)
	statFileBackup, err := os.Stat(pathToBackup)
	if err != nil {
		return err
	}
	modTimeFileBackup := statFileBackup.ModTime()
	log.Printf("Backup file info	UNIX: %d, Time: %v, Location: %s",
		modTimeFileBackup.Unix(), modTimeFileBackup, modTimeFileBackup.Location())
	log.Printf("Trying auth to WebDav server: %s, user: %s\n", currTask.Host, currTask.User)
	wdServer := wd.NewClient(currTask.Host, currTask.User, currTask.Password)
	arcDir := currTask.ArcDir

	//Find Arc dir on WebDav, create it if not found
	err = CreateRemoteArcDirIfNotExists(currTask, wdServer)
	if err != nil {
		return err
	}

	//Searching last file and max date on archive directory
	arcLastDate, err := GetArcLastDate(currTask, wdServer)
	if err != nil {
		return err
	}

	//Coerce the last date of the backup file from WebDav archive to a local TimeZone
	arcLastDate = arcLastDate.In(modTimeFileBackup.Location())
	log.Printf("inf doBackup\tActual arcLastDate in local TimeZone: %s", arcLastDate)
	//Making new filename to backup, if current file is newest
	if modTimeFileBackup.Unix() > arcLastDate.Unix() {
		curExt := filepath.Ext(currTask.FileName)
		timeToName := modTimeFileBackup.Format("02-01-2006_15.04.05-0700")
		nameWithoutExt := currTask.FileName[:len(currTask.FileName)-len(curExt)]
		newFileName := nameWithoutExt + "_" + timeToName + curExt
		log.Printf("Name without extention: %s", nameWithoutExt)
		log.Printf("New Filename to backup file: %s", newFileName)
		bytes, _ := ioutil.ReadFile(pathToBackup)
		pathToNewBackupFile := path.Join(arcDir, newFileName)
		log.Printf("Uploading backup file to WebDev path: %s", pathToNewBackupFile)
		err := wdServer.Write(pathToNewBackupFile, bytes, 0644)
		if err != nil {
			return err
		}
		log.Printf("Copy successed")
	}
	return nil
}

func CreateRemoteArcDirIfNotExists(currTask ConfTask, wdServer *wd.Client) error {
	arcDir := currTask.ArcDir
	//Get all files and directories on WebDav root
	filesRoot, err := wdServer.ReadDir("/")
	if err != nil {
		return err
	}
	//Count item on WebDav root
	cnt := len(filesRoot)
	log.Printf("Count item on WebDav root directory: %d", cnt)
	//Listing filesRoot and directories, find arch directory
	for _, file := range filesRoot {
		cnt -= 1
		if (file.Name() == arcDir) && (file.IsDir() == true) {
			log.Printf("Archive directory \"%s\" found", arcDir)
			break
		} else if cnt == 0 {
			log.Printf("Archive directory \"%s\" NOT found", arcDir)
			//Join root WebDav path and archive directory
			arcUrl, err := url.Parse(currTask.Host)
			if err != nil {
				return err
			}
			arcUrl.Path = path.Join(arcUrl.Path, arcDir)
			arcUrlStr := arcUrl.String()
			log.Printf("Creating archive directory on path: %s", arcUrlStr)
			err = wdServer.Mkdir(arcDir, 700)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
