package filter

import (
	"common"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"regexp"
	"sort"
	"strings"
)

type Os struct {
	Id				 string
	Os_type          string
	Os_version       string
	Os_keyword       string
	Os_pattern       string
	Os_patterntype   string
	Priority         int
	Keyword_priority int
}

type Osfilter struct {
	Oskeyword  common.SortQueue
	Osinfo     map[string]*common.SortQueue
	regexobj   map[string]*regexp.Regexp
	Nextfilter Filter
}

func NewOsfilter() (sf *Osfilter) {
	sf = new(Osfilter)
	sf.Oskeyword = make(common.SortQueue, 0)
	sf.Osinfo = make(map[string]*common.SortQueue)
	sf.regexobj = make(map[string]*regexp.Regexp)
	sf.Nextfilter = nil
	return
}

func (self *Osfilter) ReadInfo(connstring string) error {
	db, err := sql.Open("mysql", connstring)
	if err != nil {
		return err
	}
	defer db.Close()

	sql := "select id, os_type, os_version, os_keyword, os_pattern, os_patterntype, priority, keyword_priority from conf_os_info order by os_keyword"
	rows, _ := db.Query(sql)
	for rows.Next() {
		tmp := &Os{}
		rows.Scan(&tmp.Id, &tmp.Os_type, &tmp.Os_version, &tmp.Os_keyword, &tmp.Os_pattern, &tmp.Os_patterntype, &tmp.Priority, &tmp.Keyword_priority)

		if queue, ok := self.Osinfo[tmp.Os_keyword]; ok {
			item1 := &common.SortItem{tmp, fmt.Sprintf("%03d", tmp.Priority)}
			*queue = append(*queue, item1)//???
		} else {
			//oskeyword中新增关键字
			item := &common.SortItem{tmp.Os_keyword, fmt.Sprintf("%03d", tmp.Keyword_priority)}
			self.Oskeyword = append(self.Oskeyword, item)

			//创建新的SortQueue
			item1 := &common.SortItem{tmp, fmt.Sprintf("%03d", tmp.Priority)}
			t := make(common.SortQueue, 0)

			t = append(t, item1)
			self.Osinfo[tmp.Os_keyword] = &t
		}

		if tmp.Os_patterntype == "regex" {
			obj, err := regexp.Compile(tmp.Os_pattern)
			if err == nil {
				self.regexobj[tmp.Os_pattern] = obj
			}
		}
	}
	sort.Sort(self.Oskeyword)
	for _, v := range self.Osinfo {
		sort.Sort(*v)
	}
	return nil
}

func (self *Osfilter) GetOsinfo(uastr string) (string, string, string) {
	for _, v := range self.Oskeyword {
		oskeyword, ok := v.Value.(string)
		if ok && strings.Contains(uastr, oskeyword) {
			queue, ok := self.Osinfo[oskeyword]
			if !ok {
				continue
			}
			for _, qv := range *queue {
				os, ok := qv.Value.(*Os)
				if !ok {
					continue
				}
				if os.Os_patterntype == "str" {
					if strings.Contains(uastr, os.Os_pattern) {
						return os.Os_type, os.Os_version,os.Id
					}
				} else if os.Os_patterntype == "regex" {
					regexobj, ok := self.regexobj[os.Os_pattern]
					if ok && regexobj.FindStringIndex(uastr) != nil {
						return os.Os_type, os.Os_version, os.Id
					}
				}
			}
		} else {
			continue
		}
	}
	return "", "", "-1"
}

func (self *Osfilter) Filt(tmp map[string]string) {
	issuccess := false
	v, ok := tmp["ua"]
	if ok {
		if v == "" || v == "-" {
			tmp["os_type"] = "-"
			tmp["os_version"] = "-"
			tmp["os_conf_id"] = "-1"
		} else {
			os_type, os_version ,os_id := self.GetOsinfo(v)
			if os_type != "" && os_version != "" {
				tmp["os_type"] = os_type
				tmp["os_version"] = os_version
				tmp["os_conf_id"] = os_id
			} else {
				tmp["os_type"] = "-"
				tmp["os_version"] = "-"
				tmp["os_conf_id"] = "-1"
			}
		}
	}

	if !issuccess && self.Nextfilter != nil {
		self.Nextfilter.Filt(tmp)
	}
}
