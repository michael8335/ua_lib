// terminal_check project main.go
//读取收到的pn包，解析终端信息
//author：YangFei
package main

import (
	"bufio"
	"database/sql"
	"filter"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/robfig/config"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type checker_configs struct {
	filepath   string
	outputpath string
	workernum  int
	connstring string
}

type uaquery struct {
	ua         string
	resultchan chan bool
}

type uaresult struct {
	ua      string
	uaparam string
}

func uaqueryroutine(uacache map[string]string) chan<- *uaquery {
	querychan := make(chan *uaquery, 1024)
	go func() {
		for {
			tmp := <-querychan
			if _, ok := uacache[tmp.ua]; ok {
				tmp.resultchan <- true
			} else {
				tmp.resultchan <- false
			}
		}
	}()
	return querychan
}

func uareiveroutine(uacache map[string]string) chan<- *uaresult {
	receivechan := make(chan *uaresult, 1024)
	go func() {
		for {
			tmp := <-receivechan
			uacache[tmp.ua] = tmp.uaparam
		}
	}()
	return receivechan
}

func reader(inputfile string, filter filter.Filter, querychan chan<- *uaquery, receivechan chan<- *uaresult, readerwg *sync.WaitGroup) {
	defer readerwg.Done()
	istream, err := os.Open(inputfile)
	if err != nil {
		fmt.Println("Failed to open the input file:", inputfile)
		return
	}
	defer istream.Close()

	br := bufio.NewScanner(istream)
	br.Split(bufio.ScanLines)

	resultchan := make(chan bool)

	for br.Scan() {
		line := br.Text()

		t := strings.Split(line, "\t")
		if len(t) < 4 {
			continue
		}
		querydata := &uaquery{t[2], resultchan}
		querychan <- querydata
		uacached := <-resultchan
		if uacached {
			continue
		}
		tmp := map[string]string{"ua": t[2]}
		filter.Filt(tmp)
		result := &uaresult{tmp["ua"], fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			tmp["ua"], tmp["brand_eng"], tmp["ua_name"],tmp["ua_conf_id"], tmp["os_type"],
			tmp["os_version"],tmp["os_conf_id"], tmp["browser_type"], tmp["browser_version"], tmp["browser_conf_id"])}
		receivechan <- result
	}
}

func writer(outputfile string, result map[string]string) {
	ostream, err := os.Create(outputfile)
	if err != nil {
		fmt.Println("Failed to open the output file:", outputfile)
		return
	}
	defer ostream.Close()

	for _, tmp := range result {
		ostream.WriteString(tmp)
	}
}

func loader(outputfile string, connstring string) (sql.Result, error) {
	db, err := sql.Open("mysql", connstring)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	//create empty table
	sqlstring := "create table if not exists t_ua_library ( " +
		"ua varchar(256), " +
		"brand_eng varchar(50), " +
		"ua_name varchar(50), " +
		"ua_conf_id int(32), " +
		"os_type varchar(50), " +
		"os_version varchar(50), " +
		"os_conf_id int(32), " +
		"browser_type varchar(50), " +
		"browser_version varchar(50), " +
		"browser_conf_id int(32), " +
		"ua_date timestamp default now(), " +
		"primary key (ua)" +
		") ENGINE=MyISAM DEFAULT CHARSET=utf8;"
	r, err := db.Exec(sqlstring)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	mysql.RegisterLocalFile(outputfile)
	sqlstring = "load data local infile '" +
		outputfile + "' replace into table t_ua_library " +
		"fields terminated by '\\t' " +
		"(ua, brand_eng, ua_name, ua_conf_id, os_type, os_version, os_conf_id, browser_type, browser_version, browser_conf_id)"
	println(sqlstring)
	r, err = db.Exec(sqlstring)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return r, err
}

func readconfig() (*checker_configs, error) {
	c, err := config.ReadDefault("config.cfg")
	if err != nil {
		fmt.Printf("read config failed: %v\n", err)
		return nil, err
	}

	OUTPUTPATH, err1 := c.String("terminal_check", "OUTPUTPATH")
	if err1 != nil {
		fmt.Printf("read config failed: %v\n", err)
		return nil, err1
	}
	FILEPATH, err2 := c.String("terminal_check", "FILEPATH")
	if err2 != nil {
		fmt.Printf("read config failed: %v\n", err)
		return nil, err2
	}
	WORKERNUM, err3 := c.Int("terminal_check", "WORKERNUM")
	if err3 != nil {
		fmt.Printf("read config failed: %v\n", err)
		return nil, err3
	}
	CONNSTRING, err4 := c.String("terminal_check", "CONNSTRING")
	if err4 != nil {
		fmt.Printf("read config failed: %v\n", err)
		return nil, err4
	}
	return &checker_configs{FILEPATH, OUTPUTPATH, WORKERNUM, CONNSTRING}, nil
}

func main() {
	//Read config
	cfg, cfgerr := readconfig()
	if cfgerr != nil {
		return
	}
	runtime.GOMAXPROCS(cfg.workernum)

	fmt.Printf("%v\n", time.Now())
	err := os.MkdirAll(cfg.outputpath, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	uafilter := filter.NewUafilter()
	osfilter := filter.NewOsfilter()
	browserfilter := filter.NewBrowserfilter()
	uafilter.ReadInfo(cfg.connstring)
	osfilter.ReadInfo(cfg.connstring)
	browserfilter.ReadInfo(cfg.connstring)

	uafilter.Nextfilter = osfilter
	osfilter.Nextfilter = browserfilter
	//获取当天pn目录
	fpath :=cfg.filepath+time.Now().Format("20060102")+"/*"
	fmt.Println(fpath)  
	filelist, err := filepath.Glob(fpath)
	if err != nil {
		fmt.Printf("error:%v\n", err)
	}

	uaresultcache := make(map[string]string, 1000)
	uaquery := uaqueryroutine(uaresultcache)
	uareceive := uareiveroutine(uaresultcache)

	readerwg := new(sync.WaitGroup)

	for _, v := range filelist {
		if strings.Contains(v, "seek") {
			continue
		}
		//the line below is very important
		readerwg.Add(1)
		go reader(v, uafilter, uaquery, uareceive, readerwg)
	}
	readerwg.Wait()

	outputfile := fmt.Sprintf("%s/terminal_result.data", cfg.outputpath)
	writer(outputfile, uaresultcache)
	loader(outputfile, cfg.connstring)
	fmt.Printf("%v\n", time.Now())
}
